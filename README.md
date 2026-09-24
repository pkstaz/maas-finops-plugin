# MaaS FinOps

A community plugin for the **Red Hat OpenShift AI (RHOAI) Dashboard** that adds showback and chargeback for **Models-as-a-Service**.

RHOAI 3.x does not bill in dollars: subscriptions expose `tokenRateLimits` and the built-in Usage view is showback only. This plugin fills that gap:

- Discovers models (`LLMInferenceService`, `MaaSModelRef`, external models)
- Tokens and requests per **subscription** and per **API key / user** (Limitador metrics)
- Catalog of **price per million tokens** (ConfigMap, not a product field)
- Estimated cost = `tokens / 1e6 × price`

The frontend is a Webpack 5 Module Federation remote loaded by the dashboard. The Go BFF aggregates Kubernetes inventory, Thanos metrics, and the pricing catalog.

## Pages

| Page | What it shows |
|---|---|
| Overview | Tokens, cost, requests, 429s, top models |
| Models | Cluster models plus editable price per token |
| Subscriptions | `MaaSSubscription` usage and cost |
| API keys | Limitador `user` label (API key owner) |
| Pricing | Azure / AWS / IBM Cloud hardware, recommended $/token, manual cost |
| Pricing Simulator | What-if GPU SKU × Red Hat model token price |

## Deploy on an existing dashboard

Prerequisites: Helm, `oc` CLI access, and access to `redhat-ods-applications` (typically cluster-admin). MaaS must be installed.

### 1. Install the plugin

```bash
helm install maas-finops chart/ \
  --namespace cp-maas-finops \
  --create-namespace \
  --set image.tag=0.1.1 \
  --set bff.image.tag=0.1.1 \
  --set adminUser=admin
```

To use images from your own registry:

```bash
helm install maas-finops chart/ \
  --namespace cp-maas-finops \
  --create-namespace \
  --set image.repository=quay.io/<org>/maas-finops \
  --set image.tag=0.1.1 \
  --set bff.image.repository=quay.io/<org>/maas-finops-bff \
  --set bff.image.tag=0.1.1 \
  --set adminUser=admin
```

This creates a frontend Deployment/Service (`maas-finops`, Nginx on port 8080) and a BFF Deployment/Service (`maas-finops-bff`, Go on port 3000).

### 2. Register with the RHOAI Dashboard

```bash
oc get configmap federation-config \
  -n redhat-ods-applications \
  -o jsonpath='{.data.module-federation-config\.json}' \
| python3 -c "
import json, sys
config = json.load(sys.stdin)
config.append({
  'name': 'maasFinops',
  'backend': {
    'remoteEntry': '/remoteEntry.js',
    'authorize': False,
    'tls': False,
    'service': {
      'name': 'maas-finops',
      'namespace': 'cp-maas-finops',
      'port': 8080
    }
  },
  'proxyService': [{
    'path': '/maas-finops/api',
    'pathRewrite': '/api',
    'authorize': True,
    'tls': False,
    'service': {
      'name': 'maas-finops-bff',
      'namespace': 'cp-maas-finops',
      'port': 3000
    }
  }]
})
print(json.dumps(config))
" > /tmp/mf-config-extended.json

oc set env deployment/rhods-dashboard \
  -n redhat-ods-applications \
  "MODULE_FEDERATION_CONFIG=$(cat /tmp/mf-config-extended.json)"
```

Reload the RHOAI dashboard after the dashboard pods roll out. The plugin appears under **Community plugins → MaaS FinOps**.

### 3. Verify

```bash
oc set env deployment/rhods-dashboard -n redhat-ods-applications --list \
  | grep '^MODULE_FEDERATION_CONFIG=' \
  | head -n1 \
  | python3 -c "import json,sys; d=json.loads(sys.stdin.read().split('=',1)[1].strip()); print('\n'.join(e['name'] for e in d))"
```

You should see `maasFinops` in the list.

Full Helm customization, BFF details, and uninstall: [Deploying on OpenShift](docs/deployment/OPENSHIFT_DEPLOY.md).

## Local development

Developing a dashboard plugin is easiest with a running RHOAI dashboard connected to a real OpenShift cluster. See [Local Setup](docs/development/LOCAL_SETUP.md).

Once the dashboard is running:

```bash
npm install
npm run start:dev          # plugin dev server on port 9500
```

Start the BFF (mock data, no cluster required):

```bash
make dev-bff               # Go BFF on port 3000 with MOCK_DATA=true
```

Against a live cluster instead:

```bash
cd bff
DEV_MODE=true PORT=3000 go run ./cmd/backend
```

### Build and test

```bash
npm run build           # Production frontend build to dist/
npm test                # Frontend tests
npm run lint            # ESLint on src/ + markdownlint on **/*.md
make test-bff           # Go tests
make validate           # Frontend + BFF
```

## Metrics

Thanos HTTP (`:10902`) in `redhat-ods-monitoring`:

```text
sum by (model_name) (increase(vllm:prompt_tokens_total[24h]))
sum by (model_name) (increase(vllm:generation_tokens_total[24h]))
sum by (model) (increase(authorized_hits_total[24h]))
sum by (subscription) (increase(authorized_hits_total[24h]))
sum by (user,subscription,model) (increase(authorized_hits_total[24h]))
```

Input/output come from vLLM. Totals by subscription/API key still use Limitador `authorized_hits` (`usage.total_tokens` from the gateway).

## Pricing

The ConfigMap `maas-finops-pricing` stores:

```json
{
  "currency": "USD",
  "defaultPricePerMillion": 0.15,
  "models": {
    "qwen3-06b": { "displayName": "Qwen/Qwen3-0.6B", "pricePerMillion": 0.05 }
  }
}
```

Only users bound to `maas-finops-admin` can write the catalog from the UI (`--set adminUser=<user>`).

Recommended hosted-model price: `(USD/hour × (1+margin)) / (tok/s × 3600 × utilization) × 1e6`. Default utilization is **80%**.

## Cluster requirements

- OpenShift with RHOAI 3.3+ dashboard
- MaaS (`models-as-a-service`)
- MaaS observability (Limitador scrape + TelemetryPolicy)
- For API-key attribution: Limitador `user` label (`captureUser` or the usage patch)

If Thanos is unreachable the BFF still starts; local `MOCK_DATA=true` serves demo series.

## Catalog submission

This repository follows the [rh-ai-community-plugins](https://github.com/rh-ai-community-plugins/charter) spec (`plugin.yaml`, Helm chart, Apache-2.0). To list it in the catalog, open a PR against [charter/plugins.yaml](https://github.com/rh-ai-community-plugins/charter/blob/main/plugins.yaml). See [CONTRIBUTING.md](CONTRIBUTING.md).

## Documentation

- [Architecture](docs/architecture/) — plugin system internals and the BFF pattern
- [Development](docs/development/) — local setup, APIs, project layout
- [Deployment](docs/deployment/) — Helm install and dashboard registration

## License

Apache-2.0

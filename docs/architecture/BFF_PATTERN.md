# BFF (Backend For Frontend) Pattern

This plugin uses a Go BFF to compute MaaS token usage and cost. The dashboard proxies `/maas-finops/api/*` to that service.

## When to use a BFF

- Server-side aggregation of Kubernetes, Thanos, and ConfigMap data
- Credentials and cloud price lookups that must not live in the browser
- Business logic that is too heavy for the UI (SKU catalogs, recommendations)

Simple CRUD on a single Kubernetes object can use the dashboard `/api/k8s/*` pass-through instead. This plugin does not: FinOps needs cluster-wide inventory plus Prometheus.

## Token flow

```text
Browser                    Dashboard Backend              Plugin BFF              Cluster
  |                              |                            |                     |
  |-- fetch('/maas-finops/api/overview') -------------------->|                     |
  |                    [proxyService, authorize: true]         |                     |
  |                              |-- GET /api/overview         |                     |
  |                              |   Authorization: Bearer <user-token>
  |                              |--------------------------->|                     |
  |                              |                            |-- TokenReview / SAR |
  |                              |                            |-- list MaaS CRDs    |
  |                              |                            |-- query Thanos      |
  |                              |<-- JSON overview -----------|                     |
  |<-- JSON response ------------|                            |                     |
```

1. The frontend calls `/maas-finops/api/...` on the dashboard origin.
2. The dashboard matches `proxyService.path` and rewrites it to `/api/...`.
3. With `authorize: true`, the user's access token is forwarded as `Authorization: Bearer`.
4. The BFF authenticates the user (TokenReview) and uses its ServiceAccount for cluster reads. Catalog writes require the `admin` verb on `finops.maas.io/plugins`.

## Dashboard proxy configuration

```json
{
  "name": "maasFinops",
  "backend": {
    "remoteEntry": "/remoteEntry.js",
    "service": { "name": "maas-finops", "namespace": "cp-maas-finops", "port": 8080 }
  },
  "proxyService": [{
    "path": "/maas-finops/api",
    "pathRewrite": "/api",
    "authorize": true,
    "tls": false,
    "service": { "name": "maas-finops-bff", "namespace": "cp-maas-finops", "port": 3000 }
  }]
}
```

## Implementation

```text
bff/
  cmd/backend/main.go       # HTTP server, default port 3000
  pkg/server/               # handlers, k8s inventory, Thanos, pricing, simulator
  Containerfile             # UBI9 Go build, non-root runtime
```

Notable endpoints:

| Path | Purpose |
|---|---|
| `GET /api/health` | Liveness / readiness |
| `GET /api/auth/me` | Authenticated user and admin flag |
| `GET /api/overview` | Tokens, cost, top models |
| `GET /api/models` | Model inventory + usage |
| `GET /api/subscriptions` | MaaS subscriptions + usage |
| `GET /api/apikeys` | Usage by Limitador `user` label |
| `GET/PUT /api/pricing` | Pricing ConfigMap |
| `GET /api/recommend` | Suggested $/1M from GPU cost |
| `POST /api/recommend/apply` | Write recommended prices |
| `GET /api/simulator` | Hardware / model catalog |
| `GET /api/simulator/quote` | What-if token price |

Local mock mode: `DEV_MODE=true MOCK_DATA=true PORT=3000`.

## Deployment

Helm deploys `maas-finops-bff` (Go, port 3000) next to the frontend. Both are gated by `.Values.bff.enabled` (default `true`). The BFF ServiceAccount is bound to ClusterRoles that list MaaS CRDs, nodes, Infrastructure, and `cluster-monitoring-view`.

## Local development

```bash
make dev-bff
```

Or against a live cluster using your kubeconfig:

```bash
cd bff
DEV_MODE=true PORT=3000 go run ./cmd/backend
```

The webpack dev server proxies `/maas-finops/api` to `http://localhost:3000/api`. See [LOCAL_SETUP.md](../development/LOCAL_SETUP.md).

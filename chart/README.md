# maas-finops Helm chart

Deploys the MaaS FinOps RHOAI Dashboard plugin:

- Frontend Deployment/Service serving `remoteEntry.js` on port 8080
- Go BFF Deployment/Service on port 3000
- ServiceAccount, ClusterRoles, and a pricing ConfigMap

## Install

```bash
helm install maas-finops chart/ \
  --namespace cp-maas-finops \
  --create-namespace \
  --set adminUser=admin
```

Override the image registry while waiting for `quay.io/rh-ai-community-plugins` access:

```bash
helm install maas-finops chart/ \
  --namespace cp-maas-finops \
  --create-namespace \
  --set image.repository=quay.io/<org>/maas-finops \
  --set image.tag=0.1.0 \
  --set bff.image.repository=quay.io/<org>/maas-finops-bff \
  --set bff.image.tag=0.1.0
```

See [docs/deployment/OPENSHIFT_DEPLOY.md](../docs/deployment/OPENSHIFT_DEPLOY.md) for dashboard registration.

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
  --set image.tag=0.1.2 \
  --set bff.image.tag=0.1.2 \
  --set adminUser=admin
```

Override the image registry:

```bash
helm install maas-finops chart/ \
  --namespace cp-maas-finops \
  --create-namespace \
  --set image.repository=quay.io/<org>/maas-finops \
  --set image.tag=0.1.2 \
  --set bff.image.repository=quay.io/<org>/maas-finops-bff \
  --set bff.image.tag=0.1.2
```

See [docs/deployment/OPENSHIFT_DEPLOY.md](../docs/deployment/OPENSHIFT_DEPLOY.md) for dashboard registration.

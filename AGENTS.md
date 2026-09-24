# AGENTS.md

Guidance for coding agents working in this repository.

## Project Overview

This is `maas-finops`, a community plugin for the **Red Hat OpenShift AI (RHOAI) Dashboard**. It uses Webpack 5 Module Federation to expose remote modules that the dashboard loads at runtime. The BFF is a Go service that computes MaaS token usage and cost.

## Build and development commands

```bash
npm run start:dev     # Dev server on port 9500 with HMR
npm run build         # Production build to dist/
npm test              # Run all tests (Jest + jsdom)
npm run test:watch    # Watch mode
npm run test:coverage # Tests with coverage report
npm run lint          # ESLint on src/ + markdownlint on **/*.md
make validate         # Frontend typecheck + lint + test, plus Go vet/test
```

### BFF service commands

```bash
make dev-bff          # Mock-data BFF on port 3000
cd bff && DEV_MODE=true PORT=3000 go run ./cmd/backend   # Live cluster (uses kubeconfig)
cd bff && go test ./...
cd bff && go vet ./...
```

## Architecture

### Module Federation

Configured in `config/webpack.common.js`. Exposed remotes:

- `./extensions` (`src/rhoai/extensions.ts`) — area, community-plugins section, MaaS FinOps section, six nav items, wildcard route `/maas-finops/*`
- `./Icon` (`src/app/components/MaasFinopsNavIcon.tsx`)

Shared singletons: react, react-dom, react-router-dom, @patternfly/react-core, @openshift/dynamic-plugin-sdk.

Do not remove `CommunityBanner` or the shared `community-plugins` section.

### Pages

Routed under `/maas-finops/*`. All FinOps pages call the BFF at `/maas-finops/api/*` (rewritten to `/api/*` by the dashboard proxy or webpack-dev-server).

### BFF

`bff/` is a Go service (`cmd/backend`, `pkg/server`). It uses the in-cluster ServiceAccount for MaaS CRDs, nodes, Infrastructure, Thanos, and the pricing ConfigMap. The dashboard forwards the user Bearer token; TokenReview + SAR gate admin writes.

## Conventions

- Apache-2.0
- `plugin.yaml` at repo root is the catalog + Module Federation source of truth
- Helm only for deploy (`chart/`)
- Non-root containers, UBI9 bases, port 8080 (frontend) / 3000 (BFF)
- Frontend tests: `*.spec.ts` / `*.spec.tsx`
- Go tests live next to the packages they cover

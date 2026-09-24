# Project Layout

Directory structure of the plugin.

```text
.
├── src/
│   ├── index.ts                     # Webpack entry — dynamic import to bootstrap.tsx
│   ├── bootstrap.tsx                # React 18 root render (async bootstrap required by Module Federation)
│   ├── rhoai/                       # [DASHBOARD INTEGRATION] — what the host loads
│   │   ├── extensions.ts            #   Extension declarations (area, nav sections, nav items, route)
│   │   └── CommunityNavIcon.tsx     #   [SHARED] Sidebar icon for the community-plugins section — do not modify
│   └── app/                         # [PLUGIN CODE]
│       ├── App.tsx                  #   Router + CommunityBanner layout
│       ├── components/              #   Shared UI components
│       │   ├── CommunityBanner.tsx  #     [SHARED] "Community Plugin" banner — do not modify
│       │   ├── CommunityBanner.css  #     [SHARED] Banner styles — do not modify
│       │   ├── MaasFinopsNavIcon.tsx
│       │   └── FinOpsPage.tsx       #     Page chrome (title, range toggle, alerts)
│       ├── pages/                   #   One file per page/route
│       │   ├── OverviewPage.tsx
│       │   ├── ModelsPage.tsx
│       │   ├── SubscriptionsPage.tsx
│       │   ├── ApiKeysPage.tsx
│       │   ├── PricingPage.tsx
│       │   └── SimulatorPage.tsx
│       └── utils/                   #   BFF API client and formatters
├── config/                          # Webpack configs
│   ├── webpack.common.js            #   Module Federation setup, loaders, path alias (~ → src)
│   ├── webpack.dev.js               #   Dev server (port 9500), proxy rules
│   └── webpack.prod.js              #   Production build to dist/
├── bff/                             # Go Backend-For-Frontend
│   ├── cmd/backend/main.go
│   ├── pkg/server/                  #   Inventory, Thanos, pricing, simulator
│   └── Containerfile
├── chart/                           # Helm chart for OpenShift deployment
├── Makefile                         # Build, test, image, and chart targets (run `make help`)
├── plugin.yaml                      # Plugin metadata for the community catalog
├── Containerfile                    # Frontend container (Nginx)
└── bff/Containerfile                # BFF container (Go on UBI9)
```

## Codebase orientation

1. **Read** `src/rhoai/extensions.ts` — this is what the dashboard loads. It defines nav items and routes.
2. **Add pages** under `src/app/pages/` and corresponding nav entries in `extensions.ts`.
3. **Call the BFF** via `src/app/utils/api.ts` (`/maas-finops/api/...`).

## Shared vs plugin-specific

Files marked `[SHARED]` are common to all community plugins. Do not rename, remove, or modify them:

| File | Purpose |
|---|---|
| `src/rhoai/CommunityNavIcon.tsx` | Common sidebar icon for the community-plugins nav section |
| `src/app/components/CommunityBanner.tsx` | "Community Plugin" banner displayed on every page |
| `src/app/components/CommunityBanner.css` | Styles for the banner |
| `communityPluginsSectionExtension` in `extensions.ts` | Shared nav section that groups all community plugins |

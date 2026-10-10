# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.5] - 2026-10-09

### Fixed

- Tokenomics reference costs were $0.00 when the in/out token split is unknown (Limitador total only): the total is now priced at the blended 3:1 input:output rate of each reference model.

### Changed

- Published `quay.io/cestayg/maas-finops:0.1.5` and `quay.io/cestayg/maas-finops-bff:0.1.5` and synced Helm install documentation to that tag.

## [0.1.4] - 2026-10-09

### Changed

- Tokenomics comparison now prices the consumed tokens at the public list of the reference as-a-service models only: Claude Opus 4.8 (Anthropic), Gemini Flash (Google AI), and GPT-5 (OpenAI), instead of matching every model against the full provider catalog.
- Published `quay.io/cestayg/maas-finops:0.1.4` and `quay.io/cestayg/maas-finops-bff:0.1.4` and synced Helm install documentation to that tag.

## [0.1.3] - 2026-10-09

### Added

- Tokenomics page comparing consumed tokens with what they would cost as-a-service at public list prices across Azure Foundry, Azure OpenAI, AWS Bedrock, OpenAI, Anthropic Claude, and GitHub Copilot.
- BFF endpoint `GET /api/tokenomics?range=` with multi-provider variant matching (normalized names, publisher prefixes, most-specific-key wins) and per-provider totals.
- Copilot per-request premium pricing ($0.04 x multiplier) alongside per-token providers.

### Changed

- Published `quay.io/cestayg/maas-finops:0.1.3` and `quay.io/cestayg/maas-finops-bff:0.1.3` and synced Helm install documentation to that tag.

## [0.1.2] - 2026-09-24

### Added

- Live AWS EC2 and IBM Cloud GPU list prices in the BFF (used for hardware cost and recommended token prices).

### Changed

- Published `quay.io/cestayg/maas-finops:0.1.2` and `quay.io/cestayg/maas-finops-bff:0.1.2` and synced Helm install documentation to that tag.

## [0.1.1] - 2026-09-23

### Added

- Published `quay.io/cestayg/maas-finops:0.1.1` and `quay.io/cestayg/maas-finops-bff:0.1.1` and synced Helm install documentation to that tag.

## [0.1.0] - 2026-09-23

### Added

- RHOAI Dashboard community plugin for MaaS FinOps (tokens, subscriptions, API keys, pricing, simulator).
- Go BFF that aggregates MaaS CRDs, Thanos metrics, and the pricing ConfigMap.
- Helm chart with frontend, BFF, RBAC, and dashboard Module Federation registration docs.

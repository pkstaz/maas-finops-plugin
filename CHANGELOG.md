# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

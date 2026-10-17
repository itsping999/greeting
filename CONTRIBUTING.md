# Contributing

Keep this Go / Kratos template lightweight, explicit, and easy to extend. The canonical project rules are linked from [AGENTS.md](AGENTS.md) and the [documentation index](docs/README.md).

## Setup and workflow

Use Go 1.27.2 or newer. Run make doctor; install protoc and the pinned generators with make init when regeneration or generated consistency checks are needed. The default configuration disables optional external services.

Read the [architecture](docs/architecture/README.md) and relevant [development sections](docs/development/README.md), plus the applicable contract before changing code. Go changes use focused checks, then make fmt and make check. Markdown-only changes use make check-docs and git diff --check. Choose additional smoke or integration checks from the [validation table](docs/development/README.md#验证选择).

## Modules and generated files

Prefer make scaffold-module PROTO=order/v1/order.proto; add CRUD=1 for CRUD stubs. make scaffold-proto creates the source and registrations without the final Wire step. Both still require real business rules, storage behavior, and acceptance tests. Keep protobuf types at the service boundary, models and ports in biz, and outgoing adapters in data.

Never hand-edit API generated files, cmd/server/wire_gen.go, or openapi.yaml. After proto or constructor changes run make all; review generated outputs with their sources. Scaffold or registration changes also require make scaffold-smoke. Detailed ownership and command definitions live in the development documents.

## Plans, documentation, and delivery

Use make task NAME=order-create for complex cross-layer work; simple fixes need no plan. It creates an indexed plan under docs/plans/ without overwriting existing work. Keep status and evidence current, then complete the [formal-document writeback](docs/plans/README.md#完成与正式文档回写).

Documentation and navigation changes must pass make check-docs and git diff --check and satisfy the [documentation standard](docs/documentation-standard.md). CI calls the same documentation gate; local results do not establish a remote CI result.

Before delivery verify that defaults still run without optional external services, new optional providers have safe cleanup, deployment-facing configuration has explicit input and environment handling, and contracts/navigation reflect the final behavior. Record actual validation, exclusions, failures, and unverified scope.

Docker and Kubernetes changes use the relevant checks in the [release/deployment documents](docs/operations/README.md). Set RUN_SMOKE=1 only when the external smoke services are prepared. Commits, pushes, deployments, and message sending remain subject to the user's authorization.

# Harpy

Harpy is a Go-based modular backend platform for single-tenant/on-premise and multi-tenant/SaaS deployments. Its business capabilities are organized as modules inside one Go codebase; API, worker, migrator, and operator are separate runtime processes composed from those modules. The standard deployment infrastructure includes PostgreSQL 18, RabbitMQ 4, and RustFS through an S3-compatible API. The Go module path is `github.com/Onurlulardan/harpy`; the Go version is 1.26.

The current repository scope is the backend capabilities implemented in source. Forms Engine, Workflow Engine, AI module, MCP server, and historical SQL migration dumps are not part of the current scope. Do not add them as assumed product capabilities.

The setup CLI is `go run ./cmd/setup`. It validates a deployment profile, prepares local configuration and secrets, writes profile files, and prints the next deployment commands. It does not start Docker or run the full lifecycle itself.

## Working protocol

Except for read-only discovery of AGENTS, INDEX, SKILL, and relevant architecture files, before changing code or documentation:

1. Read `skills/INDEX.md`.
2. Read the complete `skills/*/SKILL.md` files matching the work.
3. For work affecting multiple modules, also read the `impact-analysis` skill.
4. Read the relevant architecture material under `docs/architecture/`.
5. If scope or ownership is unclear, establish the impact and decision boundary before editing.

Do not make implementation changes without reading the required skills.

The local API and worker run as containers in `ops/local/compose.runtime.yml`. The supported full local workflow is `task dev` (alias for `local:up`). The workflow performs the required sequence: `prepare`, `infra`, `bootstrap`, `migrate`, `installation:artifacts`, `license`, `tenant-executor`, `initialize`, `services`, and `verify` (see `.taskfiles/local.yml`). Skipping `license` prevents the on-premise API from finding its `billing_license_heads` record. Running `go run ./core/cmd/api` alone is not a substitute for this sequence.

## Rules

1. Hand-written files must satisfy the `code-organization` skill and fileguard. That skill is the source of truth for line, CRUD, and filename policy.
2. A module must not access another module's tables, domain, or adapters.
3. Handlers do not contain business logic. Repositories do not make authorization decisions.
4. Application commands define transaction boundaries; HTTP middleware does not.
5. Tenant identity comes from verified actor context, not a request header.
6. Modules in the same process do not communicate through HTTP or RabbitMQ.
7. RabbitMQ is an integration transport, not a source of truth.
8. Do not add Forms Engine, Workflow Engine, AI module, MCP server, or historical SQL dumps as assumed existing scope.
9. Dynamic queries run only through the REST Query `QuerySpec`.
10. Unknown fields, operators, or nodes must not be silently accepted.
11. Generated files are not edited manually and must retain their generated provenance header.
12. Do not run git commands unless explicitly requested by the user.
13. Do not introduce temporary frameworks, storage, transports, or domain models. Migration is a separately owned core boundary with persistent state, automatically verified exit conditions, and explicit ownership.

## Module boundaries

1. Each workstream is limited to one module or one explicitly defined owner boundary. Writable paths must be stated; overlapping write ownership is not allowed.
2. Until a module is closed, only its source, module-specific HTTP contract, migration, permission implementation, and targeted tests should change. Schema SQL is authored for Harpy.
3. A module is not complete without evidence that:
   - Endpoint, application, adapter, and any required DB routine are wired.
   - Production constructors use the real owner API and concrete adapters, not test fixtures or stubs.
   - No duplicate HTTP method/path is owned by multiple modules.
   - Tenant and permission decisions are proven fail-closed by targeted tests.
   - Migration versions do not conflict with another owner.
   - Targeted module tests pass.
4. Missing owner APIs, concrete adapters, permission decisions, DB routines, migrations, or targeted tests are blockers. A module must not be closed with these gaps.
5. Do not invent permissions. A different authorization model requires an explicit decision and fail-closed tests.
6. Reopen a closed module only when targeted tests or live runtime evidence proves a required defect. New scope alone is not a reason to reopen it.

## Shared surfaces

The following surfaces remain locked while individual modules are being completed: operation manifests, authority manifests, generated OpenAPI aggregate, global permission catalog, migration manifest, archguard counters, and fileguard counters.

Do not edit these surfaces from parallel module workstreams. After the affected modules are closed, one owner performs the shared normalization pass. Do not repeatedly patch shared aggregates while a module is reopened.

## Closure sequence

1. Targeted tests for closed modules.
2. Single assembly of operation, authority, permission, and OpenAPI aggregates.
3. Single normalization of the migration manifest.
4. `fileguard` and `archguard`.
5. Data-preserving migration application and local stack verification.
6. Login, tenant selection, and IAM authorization flow.

A build, unit test, or passing static gate alone is not proof of completion. Do not call Harpy complete until the local stack is running and these flows have been exercised.

Do not close gates by updating counters or snapshots while required wiring is missing. Login and localhost acceptance are separate owner work. That acceptance work does not contain an E2E retry loop. Updating README does not close missing wiring.

## Ownership

- `AGENTS.md`: mandatory repository workflow.
- `skills/INDEX.md`: work-type to skill routing.
- `skills/*/SKILL.md`: domain-specific procedures.
- `docs/architecture/`: current architecture and module documentation.
- `README.md`: repository overview, setup paths, runtime processes, and entry points.

Avoid duplicating rules across files. Link to the file that owns the rule.

## Completion

- Relevant skill checklists are satisfied.
- No unresolved decision remains in the affected architecture.
- Targeted tests and architecture tests pass.
- `go run ./tools/fileguard` and `go run ./tools/archguard` pass.
- Tenant, permission, and migration impact is reported.
- New behavior has explicit contracts and observability.

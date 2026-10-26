# PlatformOps Freeze Impact

Owner: `platformops`.

Affected contracts: No HTTP route, request, response, application port, or
permission identifier changes. The only implementation repair removes a
forbidden PostgreSQL function option from the paired observability detail-read
migrations. Shared authority, operation, permission, OpenAPI, parity, migration
stream, release, and gate manifests remain read-only.

Affected data/migrations: Platform schema, owned by PlatformOps. The fresh and
`v1_adopted` variants of logical version `20260801012000` remain byte-equivalent
apart from their required filename suffix. They create the same three bounded
read routines and do not mutate existing rows. The version is present in the
future module stream but absent from the current generated migration manifest,
so this pre-release source correction does not rewrite an applied migration.

Tenant and permission risk: The routines require an explicit environment ID and
read only PlatformOps-owned platform tables. HTTP permission checks and current
Instance Owner/platform-operator admission remain unchanged. Negative access
tests must stay fail-closed.

Events/jobs: None. No outbox, job, retry, lease, ordering, or DLQ behavior
changes.

V1 compatibility: The paired fresh/adopted migration behavior stays equivalent;
the repair only conforms routine options to the hardened SQL subset.

Rollout and rollback: Include the corrected logical migration in the single
integration-owner manifest normalization after all modules freeze. Before that
normalization there is no runtime rollback action. A failure is fixed forward
without editing an already generated or applied migration.

Required tests: PlatformOps unit tests, migration document tests, PostgreSQL 18
fresh/adopted migration integration tests, module vet, scoped fileguard, and a
global archguard run filtered to PlatformOps.

Open decisions: None inside the module. Global authority ordering and all
generated migration/release registration remain integration-owner work.

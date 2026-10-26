# Catalog aggregate freeze handoff

Date: 2026-08-02
Status: module-local **FROZEN**; single-owner global integration pending

## Frozen boundary

- Owner boundary: `core/internal/modules/catalog/**`.
- Catalog publishes 24 canonical routes plus four reject-only retired standalone
  migration routes. The production handler now requires exactly 28 unique,
  complete method/path pairs.
- The exact component-owner directory remains Analytics, Catalog, Forms, IAM,
  and Workflow. Both HTTP and worker production constructors now require a
  non-nil and non-typed-nil Forms demo lifecycle owner.
- Existing owner freezes remain part of this aggregate: 20 direct MCP targets,
  14 AI-executable targets, and the Catalog AI Workspace owner.
- No global authority/operation/storage manifest, aggregate OpenAPI, migration
  inventory/frontier, parity snapshot, or composition file was changed.

## Defects closed before freeze

1. The handler built 28 routes but required 27 in `Ready`, so production
   Catalog construction always failed. The count is now 28 and duplicate or
   incomplete routes fail closed.
2. Production constructors accepted a missing demo owner. Demo endpoints could
   authorize but return unavailable, and worker resume could not complete demo
   install/remove. HTTP and worker constructors now reject nil and typed-nil
   demo owners.
3. The AI Workspace PostgreSQL adapter repeated `TenantScope` after binding and
   imported transaction internals from read methods. Scope is now consumed only
   at the binder/root owner boundary; the adapter remains scope-bound without
   forbidden direct imports.
4. `catalog_demo_operation_plans` had table-wide INSERT granted to runtime.
   Paired forward migration `20260802232200` revokes table INSERT and grants
   only the eight required columns.
5. `catalog_guard_purge_approval_v2()` retained PUBLIC EXECUTE, giving the
   adopter a second executable routine. Paired forward migration
   `20260802232201` removes PUBLIC EXECUTE; adopter exposure is again exactly
   one composite cutover routine and zero tables, columns, or schema CREATE.

Historical migrations were not edited.

## Migration and runtime evidence

- Platform migrations: 8 fresh + 8 adopted.
- Tenant migrations: 43 fresh + 43 adopted.
- Total embedded documents: 102; every fresh/adopted logical pair is
  byte-identical and the new versions have no repository-wide collision.
- Main PostgreSQL 18 fresh/adopted lifecycle, replay, isolation, demo atomicity,
  and AI Workspace suite: PASS (`300.120s`) on the final 43-migration set.
- Durable Forms demo handoff suite: PASS (`303.919s`) on the final set.
- Purge approval atomicity/replay/isolation suite: PASS (`304.335s`) on the
  final set.
- Protected cutover and adopter-surface suite: PASS (`122.534s`) on the final
  set.

## Local gates

- `go test ./core/internal/modules/catalog/... -count=1`: PASS.
- `go vet ./core/internal/modules/catalog/...`: PASS.
- `go run ./tools/fileguard -root core/internal/modules/catalog`: PASS.
- Route manifest test proves 28 exact unique routes and rejects a duplicate.
- Constructor tests prove complete dependencies succeed and nil/typed-nil demo
  owners fail for both HTTP and worker.
- Final full archguard remains globally red. Catalog-filtered output contains
  only frozen global-registration/supersession work listed below; the direct
  adapter transaction-import findings introduced by AI Workspace are zero.

## Single integration-owner residue

Do these once after every module is frozen; do not reopen Catalog to regenerate
global inventories.

1. API and worker composition must return the real Forms
   `CatalogDemoComponentOwner` beside the five-owner directory and pass it as
   `Demo` to Catalog. Then prove both constructors are reachable.
2. Normalize the global runtime/OpenAPI characterization from 19 to all 24
   canonical routes: add package-version dependency list and the four demo
   routes. Keep the four retired compatibility routes classified separately.
3. Add the five missing authorities once: dependency list plus demo get,
   removal-plan, install, and remove. IAM's module-local permission resolver
   already maps their exact permission/operation pairs.
4. Register the eight currently unregistered operation-bearing Catalog source
   files in the global operation fragments.
5. Add `adapters/postgres/aiworkspace/sql/tenant` to the Catalog storage owner
   claim. Normalize the API UUID rule consistently with the other published
   module APIs; do not introduce a Catalog-only exception.
6. Add tenant migrations `20260802232200..01` to the one-time global migration
   order/inventory/frontier update. Teach the global SQL guard to accept exact
   table-level INSERT revocation and to recognize the forward repair of the
   historical broad grant; never rewrite the historical migration.
7. Wire the already-frozen Catalog direct/AI MCP owners exactly once and retain
   counts 20 direct / 14 AI.
8. Through Gateway, live-test package/dependency reads, install, enable,
   disable, demo status/install/removal plan/remove, menu read/override, and
   worker resume on a real tenant.

Catalog is frozen. Reopen only for a reproducible Catalog-local defect, not for
global manifest, migration-inventory, parity, or composition regeneration.

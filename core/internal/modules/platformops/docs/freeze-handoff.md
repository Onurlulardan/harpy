# PlatformOps Freeze Handoff

Status: `FROZEN` for module-owned implementation.

Scope: Only `core/internal/modules/platformops/**` was writable. Shared authority,
operation, permission, OpenAPI aggregation, parity, migration-stream/release,
composition, and global gate files stayed read-only.

## Reproducible findings and repairs

Filtered archguard originally reported exactly two module-local failures:

1. Fresh logical migration `20260801012000` declared `STABLE` on hardened
   `SECURITY DEFINER` routines.
2. The paired `v1_adopted` migration declared the same forbidden option.

Both variants now omit `STABLE`. The three routines remain typed,
`SECURITY DEFINER`, and pinned to `search_path=pg_catalog`; public execution is
revoked before the runtime role is granted access. No route, DTO, application
port, permission ID, table, or data mutation changed.

## Completeness evidence

- Endpoint: 38 unique runtime method/path pairs; 16 Platform Operations,
  3 Database Diagnostics, and 19 Recovery/Restore operations. Those counts
  match the three module HTTP path contracts, with no duplicate route owner.
- Application: every public operations, diagnostics, recovery, restore, trace,
  alert, and permission-cache handler reaches a typed application API. The three
  new detail reads validate environment/resource identity before dispatch.
- Adapter: `runtimeStore` binds concrete PlatformOps PostgreSQL operations and
  has compile-time assertions for diagnostics, recovery, restore, trace, and
  alert ports. Detail reads use snapshot-query transactions and bounded owned
  SQL routines.
- Database/migrations: 48 fresh plus 48 adopted documents; every pair is
  byte-identical and no pair is missing. Logical version `20260801012000` has one
  owner (`platformops`) across the repository platform scope.
- Permissions: the three HTTP contracts use 17 distinct platform capabilities;
  all 17 exist in the IAM platform capability catalog. Runtime access requires
  a current verified IAM decision, expected deployment mode, and exact tenant
  equality. Recovery surfaces fail closed outside on-premise mode and protected
  mutations retain their idempotency and step-up binding.
- Constructor: `NewRuntime` rejects every missing dependency, binds the real
  PostgreSQL pool, typed services, IAM access adapter, tenancy environment
  reader, IAM permission-cache owner API, and recovery-anchor orchestration.
  It does not depend on a test fixture or stub.

## Commands run

Passed:

- `go test ./core/internal/modules/platformops/...`
- `go test -count=1 -p=1 -timeout=12m -tags=integration ./core/internal/modules/platformops/adapters/postgres/integrationtest`
  against exact PostgreSQL 18.4 image; both fresh and adopted origins applied.
- `go vet ./core/internal/modules/platformops/...`
- `go run ./tools/fileguard -root core/internal/modules/platformops`
- Repository-wide read-only migration pair/count and version-owner checks.
- Repository-wide `go run ./tools/archguard`, filtered to PlatformOps, after
  repair: no module-local finding remains.

Skipped or externally blocked:

- `go test -race ./core/internal/modules/platformops/...` cannot run in this
  Windows environment because `CGO_ENABLED=0`; the normal suite is green.
- `go test ./core/internal/composition/api` currently cannot compile because
  the out-of-scope `formsRecordApplicationFixture` does not yet implement the
  current Forms `DeleteV1Many` application method. This does not originate in
  PlatformOps, and no PlatformOps or composition file was changed.

Compatibility result: The paired fresh/adopted behavior is unchanged except for
conformance to the hardened SQL option allowlist. HTTP and permission contracts
are unchanged.

Rollback evidence: Logical version `20260801012000` is not in the current
generated migration manifest. The source pair was corrected before the single
integration-owner generation pass, and both origins pass the real PG18 gate.

## Integration-owner debt only

1. Sort `platformops.trace_spans.get` in the locked shared authority manifest;
   this is the sole remaining PlatformOps-filtered archguard result.
2. During the one permitted post-freeze normalization, generate and attest the
   corrected migration checksum/frontier/release entries without reopening this
   module.
3. Re-run the composition package after the Forms owner freezes its current
   application-interface update; no PlatformOps change is indicated by that
   compiler failure.

Do not reopen PlatformOps for parity counts or shared-manifest drift. Reopen only
if a targeted test or live Gateway flow reproduces a module-owned defect.

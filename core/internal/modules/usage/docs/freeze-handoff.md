# Usage module freeze handoff

Status: `MODULE_LOCAL_FROZEN`; do not reopen without a reproducible targeted
or live-runtime defect. Shared manifests, migration inventory and composition
remain reserved for the single integration owner.

## Exact legacy surface

The V1 `log-service` exposes eight `/api/usage-analytics` reads for the
platform scope and the same eight reads under `/api/usage-analytics/tenant`:

- overview
- users/active
- users/top
- users/{user_id}/summary
- users/{user_id}/timeline
- features
- heatmap
- trend

V2 owns the exact sixteen method/path pairs. The separate legacy
`/api/analytics` aliases belong to the frozen Analytics module and are not
double-counted as Usage capabilities.

Tenant reads require `analytics.usage.read`; SaaS control-plane reads require
`platform.analytics.usage.read`. Each route has a distinct tenant and platform
operation ID, and both the Go application and PostgreSQL routine recheck the
exact scope, permission and operation.

## Parity and intentional hardening

- Legacy response field names, default limits, 24-hour default window,
  pagination shape, overview composition, DAU/WAU/MAU, heatmap and trend
  outputs are preserved.
- Tenant results retain the actor UUID expected by tenant administrators.
  Platform results use a stable environment/tenant/actor SHA-256 pseudonym so
  a control-plane response cannot expose raw tenant or user identity.
- Legacy feature labels `Form: <slug>`, `Traces`, `System Health` and all other
  existing categories are preserved. Canonical V2 route templates and the new
  `Platform Operations` category are supported without reintroducing raw URLs.
- Invalid or duplicate query values now fail with stable `422 INPUT_INVALID`
  instead of V1's silent fallback. Windows are bounded to 31 days and platform
  user detail accepts only an issued 64-character pseudonym.
- Immutable facts store method, canonical route template, status, duration and
  time only. Query strings, raw instance paths, IP address, user agent,
  request/response bodies and tokens are intentionally not retained. Error
  rate derives from HTTP status `>= 400`, not mutable log-level text.

## Module-local closure

- All nine request-runtime PostgreSQL calls, the retention-owner call, six
  adoption calls and two legacy source reads are static embedded SQL documents.
  Go adapters contain no embedded SQL literals or dynamic relation interpolation.
- Legacy `trace_logs` adoption accepts only the exact relation name and a
  validated schema. The transaction runner sets a local, parameterized
  `search_path`; source OID, count and root hash are rechecked before every
  page, apply, advance, reconcile and completion step.
- V1 rows receive deterministic UUIDv7 target identities and durable source to
  target mappings. Page receipts, lease generation, cursor, transform digest
  and source drift fences make restart and replay explicit.
- Domain identity is a standard-library value. Google UUID conversion exists
  only at application/adapter boundaries, and integration tests no longer
  import sibling adapter packages.
- Fifteen `fresh` migrations and fifteen `v1_adopted` migrations use versions
  `20260801050000..20260801050014`; every pair is byte-identical and no version
  collides with another module stream.
- Usage tables are owner-only. Runtime and adopter roles have no direct table
  privilege; runtime receives exactly nine read/record routines, adopter
  receives exactly nine retention/adoption routines, and PUBLIC receives none.

## Controlled retention reopen closure

- `usage.retention.prune` revision 1 is the module-owned global,
  database-transactional job contribution. Its payload is only the canonical
  environment/cycle reference and SHA-256 digest; it contains no tenant, actor,
  route or request data.
- The owner claim accepts only exact `global/<environment>` scope, `usage`
  group/domain, V2 writer origin, positive live group/domain/lease generations,
  UUIDv7 job and attempt identities, and a due canonical cycle. Zero,
  cross-environment or tampered evidence fails before owner SQL is reached;
  the global transaction store remains responsible for proving that positive
  generations are current under lock before it constructs this evidence.
- The owner pins raw-fact retention to 400 days and a fixed 1,000-row batch.
  The job ID is the prune receipt ID, so a retry returns the first durable
  count. A full batch produces an immediate bounded successor; an idle cycle
  advances to the next future daily anchor without replaying missed empty days.
- `RetentionJobBinding` binds the Usage PostgreSQL owner inside a caller-owned
  transaction. `RetentionJobProcessor` exposes one exact global store port whose
  contract requires live authority verification plus prune effect, successor
  enqueue and job completion in one XID. Composition therefore never imports or
  embeds `usage_retention_prune_v1` SQL.
- The adapter invokes only the existing SECURITY DEFINER prune routine. Legal
  holds, environment filtering, immutable-fact delete guard, bounded deletion
  and receipt conflict checks remain enforced by PostgreSQL for both origins.
  No migration, global catalog, composition, config or manifest changed in this
  controlled reopen.

## Verification evidence

- `go test ./core/internal/modules/usage/... -count=1` -- pass.
- `go vet ./core/internal/modules/usage/...` -- pass.
- `go run ./tools/fileguard -root core/internal/modules/usage` -- pass.
- `go test -tags=integration ./core/internal/modules/usage/adapters/postgres/integrationtest -count=1 -timeout 12m` -- pass on exact PostgreSQL 18.4 for both origins, covering migration application, ACL, all read shapes, tenant isolation, platform pseudonymization, idempotent/concurrent record replay, retention and legal holds.
- `go test -tags=integration ./core/internal/modules/usage/adapters/postgres/adoption -run '^TestUsageAdoptionOnFreshAndAdoptedPostgreSQL184$' -count=1 -timeout 12m` -- pass on exact PostgreSQL 18.4 with real legacy `trace_logs`, restart and source-drift checks.
- Global archguard still fails while locked repo-wide registrations remain
  intentionally unnormalized. The Usage-path filter has zero module-layer,
  restricted-import, dynamic-SQL, embedded-SQL or broad-permission defects.
  Its exact 57 findings are locked registrations: eight operation sources,
  eighteen static SQL documents, one migration provider and thirty migration
  files.

## Single-owner global integration debt

No item below authorizes a Usage module edit:

1. Register the eight operation-bearing sources and the sixteen exact read
   authorities once in the locked operation/authority manifests. Register the
   internal record and leased adoption authorities without inventing public
   routes.
2. Add one Usage storage owner claim for the platform migration directory,
   both static-SQL trees, canonical relations/routines and the bounded legacy
   `trace_logs` source adapter.
3. Add the paired `20260801050000..20260801050014` stream and migration
   provider once, selecting its generation only after every module is frozen.
4. Compose the profile-bound PostgreSQL command/snapshot runners, clock,
   UUIDv7 generator, IAM access verifier, service, HTTP handler and all sixteen
   routes in API composition.
5. Attach a post-response Gateway observer that records one canonical fact per
   admitted request with a stable source identity. Its failure policy and
   recursion decision must be explicit; it must never persist raw request data
   or recursively count Usage reads without an accepted decision.
6. Add the module contribution from `RetentionJobPolicies()` to the one-time
   release catalog. The current platform job schema permits global Billing only
   and the standard database processor rejects global claims; normalize the
   global cutover vocabulary/group for exact `usage` ownership and implement
   `RetentionJobStore.ProcessUsageRetentionJob` as the global atomic store.
7. That store must map the freshly locked environment scope, group/domain
   generations, V2 writer origin and lease generation into
   `NewRetentionAuthority`, then call `RetentionJobBinding.Bind` and commit the
   prune receipt, successor and completion in one transaction. The existing
   prune routine is executable only by `harpy_adopter`; the selected worker
   profile or paired bounded wrapper/grant must preserve this atomicity and may
   not grant table DML or bypass the generation checks.
8. Bootstrap one canonical recurring job per environment and register
   `RetentionJobProcessor` in the executable worker registry only after the
   catalog and global transaction store above exist.
9. Compose zero/legacy adoption under the adopter lease authority and provide
   the exact legacy schema, environment and default tenant inputs.
10. Refresh the global parity inventory once after all modules freeze. Mark the
   sixteen Usage routes exact/redesigned with this evidence; do not count the
   Analytics aliases or protocol tools as additional missing features.
11. After one-time normalization, apply migrations to the real local database
   and prove tenant and platform Usage reads through Gateway port 8000 with a
   live IAM decision and newly observed traffic.

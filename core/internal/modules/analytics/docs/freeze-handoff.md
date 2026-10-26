# Analytics Aggregate Freeze Handoff

Date: 2026-08-02

Status: **MODULE-LOCAL FROZEN; GLOBAL INTEGRATION PENDING**

Reopen `core/internal/modules/analytics/**` only for a reproducible Analytics
defect. Shared manifests, composition, aggregate OpenAPI, permission catalogs,
migration inventories and global guard snapshots remain single-integration-owner
work. No locked global integration file was edited during this freeze.

## Frozen HTTP and capability surface

The Analytics HTTP facade owns exactly 38 unique method/path entries:

- saved query CRUD, publication and execution: 10;
- query batch execution: 1;
- dashboard CRUD/default/publication: 10;
- asynchronous dashboard export: 4;
- publication and public execution: 10;
- legacy public execution aliases: 4.

Construction and `Ready` fail closed for a missing, duplicate or incomplete
route. The complete handler requires query, dashboard, publication, export,
public and compatibility applications; typed-nil dependencies are rejected.

The separately frozen MCP owner has 20 direct names and 14 executable names.
That inventory is recorded in `mcp-freeze-handoff.md` and is not inferred from
HTTP route counts.

## Legacy parity disposition

The legacy Gateway-exposed Analytics/OpenAPI/Dashboard surface also has 38
routes, but aliases are not counted as separate product features:

| Legacy capability | V2 disposition |
| --- | --- |
| Query definition aliases | Redesigned as versioned saved queries. |
| Validate aliases | Publish/revalidate exact canonical definitions. |
| Forms metadata aliases | Forms-owned capability, not Analytics duplication. |
| Slug availability | Intentional ID-first identity redesign. |
| Raw unsaved preview | Intentionally retired unsafe second query engine. |
| Batch execution | Canonical `/api/analytics/query-executions`. |
| Dashboard CRUD/default/validate | Revisioned dashboard aggregate. |
| Embed | Approved opaque publication lifecycle. |
| Synchronous CSV/XLSX/PDF | Async encrypted export handle/content contract. |
| Public execute aliases | Compatibility mapping retained. |
| Private V1 execute paths | Consumer cutover to saved-query execution. |
| Internal license-count routes | Infrastructure-only; Billing owns behavior. |

## Billing-owned mutation admission

Analytics application packages consume the public
`billing/api.MeteredResourceMutationDecider` contract directly. The tiny
Analytics composition alias owns no validation, feature, gauge or replay
semantics.

- Query/dashboard create or restore requires builder entitlement and atomic
  `+1` gauge admission.
- Update, publish and revalidate require builder entitlement and do not change
  the gauge.
- Delete remains available when the builder is disabled and performs an
  idempotent `-1` or no-op gauge decision.
- Publication create, submit, approve and activate are exposure-widening
  publish decisions for their query/dashboard owner kind.
- Read, execute, revoke and delete do not accept quota as authorization.

Every decision is made before the mutable Analytics owner write in the same
`TenantScope`. Request identity includes resource, mutation, tenant, actor,
proposed operation/owner fact, idempotency key and digest. A Billing denial,
error or mismatched decision prevents the owner store call. Billing allow plus
owner failure returns an error from the same scope. Billing and owner replay
flags and canonical owner facts must agree.

Billing's public API and constructor shape are final. Its PG18 rollback and
concurrency proof is still pending in the Billing-owned freeze; this is an
external acceptance blocker, not permission to reopen Analytics.

## Data and migration evidence

- Platform migrations: 26 fresh/adopted pairs.
- Tenant migrations: 140 fresh/adopted pairs.
- Total Analytics migration documents: 332.
- Paired byte mismatches: 0.
- Cross-repository Analytics version collisions: 0.
- No Analytics entitlement migration was added; Billing owns entitlement,
  receipts, gauge state and opening reconciliation.

PostgreSQL 18 proves both fresh and V1-adopted application plus protected
cutover denial/runtime evidence. Historical migrations were not edited.

## Defects closed before freeze

- Duplicate route entries could previously satisfy route-count readiness;
  exact completeness and method/path uniqueness are now enforced.
- Export scheduler/content dependencies were not constructor-mandatory; the
  Analytics runtime now rejects nil and typed-nil dependencies.
- V1 builder/limit enforcement was absent in V2; Billing-owned atomic mutation
  admission now guards every widening/configuration path.
- Publication activation replay identity omitted `CapabilityTTL`; the TTL is
  now part of the activation request digest and replay never regenerates
  capability material.
- Owner-store ordering, deny/error short-circuiting, canonical identity and
  replay agreement have targeted Query, Dashboard and Publication tests.
- Dashboard export render composition exposed a frozen-contract defect: a
  pinned widget carried only `query_version_id` while the query executor needed
  both root and version identities. The owner now resolves the active published
  root through a tenant-authorized SECURITY DEFINER routine, executes the exact
  pinned version, merges pinned/request parameters deterministically and derives
  row counts from the canonical query envelope. Worker composition now binds
  current pinned IAM export plus query-execute authorization, encrypted spool,
  FCART/S3 artifacts, render workload, recovery, delete/expiry processors and
  the render delegate. Targeted module and worker tests pass after re-freeze.

## Recorded gates

Run from `backend-v2`:

```text
go test ./core/internal/modules/analytics/... -count=1
go vet ./core/internal/modules/analytics/...
go test ./tools/fileguard -run Analytics -count=1
go test ./tools/archguard -run Analytics -count=1
go test -tags=integration ./core/internal/modules/analytics/adapters/postgres/integrationtest -run '^(TestAnalyticsMigrationsAgainstPostgreSQL184|TestAnalyticsPG18ProtectedCutoverFreshAndAdoptedEvidence)$' -count=1 -v
```

Recorded on 2026-08-02:

- all Analytics packages: PASS;
- Analytics vet: PASS;
- Analytics filtered fileguard and the 200-line source limit: PASS;
- PostgreSQL 18 fresh, adopted and protected-cutover suites: PASS, 53.677s;
- filtered archguard: expected FAIL only on 14 global operation claims and 12
  global storage claims; seven additional findings belong to AI observability.
  No Analytics-local import, MCP, migration-safety or source-layer finding
  remains.

## Single-owner integration residue

1. Construct `billing.NewMeteredResourceEntitlements(...)` once and inject its
   public decider into Analytics in the existing tenant transaction.
2. Complete Billing's rollback/concurrency gate. For opening active resources,
   wire the separate evidence-bound reconciliation API to an Analytics/API
   authoritative-row verifier; never infer counts from Gateway traffic.
3. Register the 14 operation claims and 12 storage claims once after all module
   freezes. Do not edit Analytics source to suppress those findings.
4. Normalize aggregate OpenAPI, parity dispositions, permissions, migration
   registry/frontier and global snapshots exactly once.
5. Start the real local stack, then prove through Gateway: login, tenant
   selection, IAM, query create/update/publish/execute/delete, dashboard and
   publication lifecycles, export content, MCP/OAuth and legacy public aliases.

Analytics being frozen is not a claim that backend-v2 is locally runnable or
globally parity-complete. Those claims require the integration and live gates
above.

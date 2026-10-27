# Analytics aggregate freeze impact

Date: 2026-08-02
Owner: `core/internal/modules/analytics/**`

## Affected contracts

- The module-local HTTP manifest contains 38 canonical or compatibility routes.
- Saved query, dashboard, publication, public execution, asynchronous dashboard
  export, legacy public execution and Analytics-owned MCP surfaces are in scope.
- Legacy evidence is the Gateway-exposed OpenAPI/Analytics/Dashboard surface;
  duplicate aliases are classified by behavior, not counted as new products.
- Aggregate OpenAPI, characterization policy, operation/authority/storage
  manifests and API composition remain locked for the final integration owner.

## Affected data and migrations

- Analytics owns tenant saved-query, dashboard, publication, adoption, job,
  export, audit, retention and reference state.
- Analytics owns its platform public-route, limiter and cutover projections.
- Historical migrations are immutable. Any repair must be a paired forward
  migration with byte-identical `fresh` and `v1_adopted` variants.
- Runtime access remains owner-bound through Analytics PostgreSQL binders; no
  other module table or adapter is read directly.

## Tenant and permission risk

- Authenticated tenant identity comes from the verified actor scope; public
  identity comes from the opaque platform publication pointer.
- Base IAM permission is mandatory and dashboard ACL cannot widen it.
- PUBLIC/CAPABILITY lifecycle, step-up, approval, generation and rate-limit
  checks stay fail-closed.
- Route uniqueness, typed-nil constructor dependencies, cross-tenant access and
  runtime table/function privileges require targeted negative evidence.
- Analytics consumes Billing's public `MeteredResourceMutationDecider`; it does
  not own a second feature, limit, gauge or replay policy. Query/dashboard
  create/restore require builder plus atomic `+1` gauge admission; update,
  publish and revalidate require the builder without changing the gauge;
  delete remains available with idempotent `-1` or no-op accounting.
- Publication create, submit, approve and activate are exposure-widening
  publish mutations and require the corresponding query/dashboard builder.
  Revoke and all reads/executions remain outside quota admission.
- Billing decides before every mutable Analytics owner write in the same tenant
  transaction. Public Billing conflict/request errors are mapped explicitly;
  unavailable, contradictory or mismatched decisions fail closed. Billing and
  owner replay evidence must agree or the transaction returns an error.

## Events and jobs

- Query revalidation, cache warm, publication repair and retention remain
  database-transactional Analytics jobs.
- Dashboard export render/delete remain external-effect jobs with durable
  artifact receipts, encrypted FCART storage and reconciliation.
- RabbitMQ is transport only and is not an Analytics state authority.

## V1 compatibility and rollback

- Legacy endpoint aliases normalize to versioned saved queries.
- Form discovery is served by the Forms owner; raw preview, slug identity and
  mutable embed tokens are intentionally not reproduced in the core.
- Legacy synchronous CSV/XLSX/PDF dashboard downloads are represented by the
  asynchronous encrypted dashboard-export handle and content contract.
- V1's two internal license-count HTTP routes are infrastructure-only, not
  product routes. Their behavior is replaced by Billing-owned atomic admission:
  query maps to API-builder/active-endpoint entitlement and dashboard maps to
  dashboard-builder/active-dashboard entitlement.
- Rollback uses a compatible V2 image and persisted cutover/reconciliation
  state; there is no production dual-write, second query engine or down
  migration.

## Required tests and open integration residue

- Run all Analytics unit/contract tests, vet, scoped fileguard and targeted
  archguard; run PostgreSQL 18 fresh and adopted apply/runtime suites.
- Prove the 38 method/path entries are complete and unique.
- The concrete export scheduler, artifact content runtime, worker store,
  pinned-query renderer, encrypted spool, recovery and HTTP export application
  are composed. The final integration owner must still normalize global
  manifests and prove the complete export lifecycle through Gateway.
- Billing now publishes the tenant-scoped atomic
  `MeteredResourceMutationDecider` API. Final composition must construct its
  concrete runtime and pass that public API to Analytics. It must not read a
  platform license snapshot or restore Gateway count middleware, because
  Billing remains the sole entitlement and gauge authority.
- Global parity disposition and aggregate registration are intentionally
  deferred until every module freezes.

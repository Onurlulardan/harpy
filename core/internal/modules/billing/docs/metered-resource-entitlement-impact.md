# Metered resource entitlement impact

## Owner

- `billing` remains the sole entitlement and usage authority.
- Analytics remains the owner of saved-query and dashboard lifecycle state.
- The Billing public API admits owner mutations inside the caller's existing
  `TenantScope`; it does not open a second transaction or call Analytics SQL.
- The implementation extends the existing Billing usage gauge engine. It does
  not introduce a second limit resolver, counter, cache or license decision.

## Affected contracts

- Add one module-public Billing mutation decision API for `api_endpoint` and
  `analytics_dashboard` resources.
- Requests carry mutation kind, verified tenant and actor, proposed UUIDv7
  operation/resource identities, idempotency key and request SHA-256.
- The first decision returns canonical identities. Same key and digest replays
  that decision; key reuse with another digest fails closed.
- `create` and `restore` require the exact builder boolean and atomically add
  one to the exact gauge. `delete` is builder-independent and atomically
  releases one. `update`, `publish` and `revalidate` require only the builder
  boolean.
- Query mapping is `api.builder.enabled` plus `api.active_endpoints`.
- Dashboard mapping is `analytics.dashboard_builder.enabled` plus
  `analytics.active_dashboards`.

## Affected data and migrations

- Scope is platform schema because Billing's canonical entitlement, usage and
  tenant authority records are fully qualified `platform.billing_*` objects.
- Add Billing-owned resource projection and immutable decision receipt storage.
  The projection records only owner kind, stable resource ID and counted state;
  Analytics business fields are not copied.
- Add paired fresh and V1-adopted forward migrations. No applied migration is
  edited and no destructive down migration is introduced.
- Runtime mutation uses Billing-owned typed routines or already-bound Billing
  usage adapters. Runtime roles receive no direct table DML.

## Tenant and permission risk

- The API accepts only a live command `TenantScope`; tenant and actor must match
  the scope exactly. Missing scope, authority, entitlement, meter period,
  offering, license runtime state or database row fails closed.
- Billing does not add an IAM permission. The Analytics application must perform
  its existing operation permission first and call Billing before its owner
  mutation in the same transaction.
- On-premise build-off disables builder decisions and meter increases. Delete
  and narrowing remain available as required by ADR-0120.

## Events and jobs

- No new event, job, provider call or network effect is introduced.
- Existing Billing threshold event behavior remains part of the usage gauge
  engine and therefore commits or rolls back with the caller transaction.

## V1 compatibility

- V1 `maxApiEndpoints`, `maxDashboards`, `apiBuilder` and `dashboardBuilder`
  map to the exact ADR-0120 keys above.
- Opening/adopted counts require evidence-bound resource registration; missing
  resource identity is not assumed to be zero and delete cannot decrement an
  unknown resource.
- Restore reopens capacity through the same feature and finite/unlimited limit
  admission as create. Evidence-bound reconciliation repairs projection state
  without allowing an unproved blind decrement.

## Rollout and rollback

- Apply paired additive Billing migrations, bind the public API once in final
  composition, then reconcile adopted Analytics resources against the already
  attested opening gauge snapshot before enabling owner writes.
- Rollback is to a compatible V2 image while additive receipt/projection rows
  remain. Gauge and owner mutation always share one transaction, so a failed
  owner write leaves neither a receipt nor a gauge delta.

## Required tests

- Finite, unlimited, builder-off and unavailable dependency decisions.
- Same-key replay, digest conflict, resource identity conflict and cross-tenant
  rejection.
- Concurrent admissions at the finite boundary and duplicate resource races.
- Create/delete/restore lifecycle, adopted opening and reconciliation repair.
- Caller rollback proves no receipt, projection or gauge delta survives.
- Paired migration equality plus PostgreSQL 18 fresh and V1-adopted behavior.

## Open decisions

- None. ADR-0019, ADR-0028, ADR-0079, ADR-0082 and ADR-0120 fix the owner,
  transaction, key mapping, build-off and fail-closed semantics.

## Controlled-reopen closure

- Billing is re-frozen after adding the public mutation/reconciliation
  contracts, application policy, existing-usage-engine bridge, three durable
  tables, five typed PostgreSQL routines and six exact fresh/V1-adopted pairs.
- Billing and Analytics package tests, Billing vet, pair equality, PG18
  fresh/adopted migration application, finite-boundary concurrency,
  caller-rollback atomicity and runtime ACL checks pass.
- No shared operation manifest, storage claim manifest, global migration
  inventory/frontier, parity snapshot or final composition file was edited.
  Those exact registrations are recorded in `freeze-handoff.md` for the single
  integration owner after all modules are frozen.

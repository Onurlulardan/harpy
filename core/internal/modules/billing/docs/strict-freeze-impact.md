# Billing strict-freeze impact

## Owner and boundary

- Owner: `billing`; one core module and one canonical PostgreSQL authority.
- Owned state: immutable offering/price/campaign/discount catalog, tenant base
  assignments, subscriptions, invoices, provider ledger, entitlements, usage,
  legal hold, retention and V1 adoption evidence.
- Global catalog state and tenant-bound state remain under the same owner but
  use distinct sealed runtime scopes. No second service, Redis authority,
  provider authority or Forms-owned quota is introduced.
- This pass may change only `core/internal/modules/billing/**`.

## Contracts and consumers

- Canonical HTTP surfaces include public catalog, tenant billing, platform
  catalog/reports/license operations and provider webhook ingress.
- The 26 characterized V1 method/path pairs remain compatibility-edge mappings
  into final Billing application ports; they are not 26 new domain features.
- Tenant reads require `billing.read`, usage reads `billing.usage.read`, tenant
  writes `billing.manage`; platform operations require exact
  `platform.billing.*` permissions. Public reads use a public global snapshot,
  not a synthetic actor or tenant.
- Forms and Documents consume only Billing APIs for atomic usage reservation or
  gauge accounting. MCP has no independent Billing feature surface.

## Data, tenancy and transactions

- Canonical tables and routines are fully qualified `platform.billing_*` and
  belong to the paired Billing platform migration stream.
- Global identity comes only from registered environment state. Tenant identity
  comes from verified actor/scope; provider identity comes from verified object
  mapping; job identity comes from persisted authority evidence.
- Tenant Billing requires the exact `billing` group with ordered domains
  `billing,forms`; missing or stale generation fails closed.
- Business mutation, audit, outbox/job and idempotency receipt commit together.
  Provider network effects run outside database transactions and finalize from
  durable receipts.
- Runtime roles use typed owner routines and must not receive direct table DML.

## Billing invariants

- Money is checked integer minor units with currency and pinned exponent
  registry; float, implicit conversion and unchecked overflow are forbidden.
- Published catalog, invoice, payment and applied-policy history is immutable.
- Missing base assignment, entitlement, subscription, counter or database
  result never means unlimited.
- Hard usage reservation is atomic with the owner mutation. Outcome-unknown
  effects retain capacity until reconciliation.
- Webhook tenant selection never trusts headers or payload metadata; secret and
  raw provider material stays outside logs/events and follows retention policy.

## V1 compatibility and rollout

- Exact/preserved: characterized public/tenant reads and selected error/wire
  shapes through the compatibility adapter.
- Redesigned: mutable catalog changes become immutable versions or retire
  commands; arbitrary status updates become explicit legal transitions.
- Intentional differences: UTC half-open periods, currency-grouped revenue,
  deterministic discount applicability/stacking, PostgreSQL usage authority,
  transaction-atomic event/audit and fail-closed limits.
- V1 catalog and tenant adoption remain bounded, receipt-backed and fenced.
  Global Billing precedes tenant `billing`; no production dual write or
  destructive down migration is introduced.

## Freeze verification plan

1. Inventory endpoint-to-permission-to-runtime-to-application-to-adapter/routine
   wiring and production constructors.
2. Compare the 26-route V1 characterization and current compatibility adapter;
   classify exact, redesigned and intentional differences without editing the
   global parity inventory.
3. Prove fresh/V1-adopted migration pairing, content equality and repository-wide
   scope/version collision absence using read-only scans.
4. Run Billing unit/contract tests, vet, scoped fileguard, filtered archguard and
   the real PostgreSQL 18 integration suites selected by Billing's own gates.
5. Repair only proven module-local defects, then freeze and hand operation,
   authority, storage, migration inventory, IAM/composition and live-gateway
   registrations to the single global integration owner.

## Locked global work

This pass does not edit global operation/authority/permission manifests,
OpenAPI aggregation, characterization/parity policy, migration generation or
release inventories, composition roots, expected counts or global gates.

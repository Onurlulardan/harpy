# Billing strict-freeze handoff

## Freeze verdict

**Module-local status: RE-FROZEN.** Billing was reopened once, under the
controlled-reopen rule, because Analytics/API owner mutations needed a shared
metered-resource entitlement authority. The Billing-owned public contract,
application, usage bridge, PostgreSQL routines, additive migrations and focused
tests are complete; no known module-local blocker remains.

Do not reopen `core/internal/modules/billing/**` unless a reproducible failure is
shown to originate inside the Billing boundary. Global manifest, composition,
parity, migration-order and release normalization belongs to one integration
owner after all modules are frozen.

This verdict does not claim that the complete product stack is release-ready.
Global gates, a clean migration application and live Gateway journeys remain
final integration work.

## Frozen boundary and runtime proof

- Billing is the sole owner of the immutable catalog, subscription, invoice,
  payment/provider ledger, entitlement, usage and retention authorities stored
  under qualified `platform.billing_*` objects.
- Public and platform state uses sealed global scope. Tenant state requires the
  exact `billing_forms` group and ordered domains `billing,forms`.
- API composition already builds the global, tenant and platform-report Billing
  runtimes, mounts canonical and V1 routes, and separately builds the platform
  license handler. The migrator consumes Billing migration documents and the
  worker builds provider, retention and subscription processors.
- HTTP handlers call Billing application ports; PostgreSQL adapters call typed,
  owner-qualified routines. Runtime roles do not acquire direct table DML.
- Mutations, idempotency evidence, audit and outbox/job intent commit together.
  Provider network effects execute outside the transaction and reconcile from
  durable receipts.
- `billing/api.MeteredResourceMutationDecider` now admits API endpoint and
  Analytics dashboard mutations inside the caller's existing `TenantScope`.
  It never opens a second transaction and never calls an owner module.
- Exact mappings are `api.builder.enabled` with `api.active_endpoints`, and
  `analytics.dashboard_builder.enabled` with
  `analytics.active_dashboards`. Create/restore are builder-and-gauge bound;
  delete is builder-independent; update/publish/revalidate are boolean-only.
- Canonical operation/resource identity, request digest, counted projection,
  gauge evidence and reconciliation evidence are durable and replay-pinned.

## HTTP and compatibility inventory

| Surface | Frozen count | Evidence |
| --- | ---: | --- |
| Canonical public | 2 | `PublicRoutes()` and exact route test |
| Canonical tenant | 7 | `TenantRoutes()` and exact route test |
| Canonical platform | 24 | `PlatformRoutes()` and exact route test |
| Canonical webhook | 1 | `WebhookRoutes()` and exact route test |
| Platform license | 3 | read, refresh and refresh-operation tests |
| V1 compatibility | 26 | characterized method/path pairs and bindings |

The 34 canonical routes, 3 license routes and 26 V1 mappings are transport
entries, not 63 independent product features. V1 evidence additionally covers
8 database observers, 8 Redis behaviours, 21 fixtures and 10 consumers; these
are contract evidence categories and must not be converted into a route-based
"missing feature" count.

Every characterized V1 route has an explicit bound handler. Compatibility
authorization is ordered fail-closed as availability decision, retirement hit,
then canonical operation.

## Permission freeze

- Tenant reads require `billing.read`; tenant mutations require
  `billing.manage`; usage reads require `billing.usage.read`.
- Platform catalog/report operations require their exact
  `platform.billing.*` permission. Public catalog reads use a public global
  snapshot rather than a synthetic tenant or actor.
- The IAM Billing resolver currently proves exactly 54 declared
  permission-operation pairs and rejects wildcard, wrong-operation and
  tenant/global substitutions.
- Platform license access remains environment-mode and capability bound; SaaS
  rejection occurs before IAM, while on-premise reads require the exact license
  capability.

IAM mappings are already present. They were verified read-only and were not
edited during this module freeze.

## V1 parity decision

| Decision | Billing behaviour |
| --- | --- |
| Exact/preserved | Characterized public and tenant reads, selected wire/error shapes, legacy identity resolution and explicit route availability |
| Redesigned | Mutable catalog edits create immutable versions; removal becomes retire; arbitrary subscription status writes become legal transition commands |
| Intentional difference | UTC half-open periods, deterministic discount applicability/stacking, currency-grouped revenue, PostgreSQL atomic usage authority, fail-closed missing state, transaction-atomic audit/outbox |
| Infrastructure-only | Legacy Redis counters and check-then-increment are replaced by owner PostgreSQL routines; this is not a missing Billing feature |
| Missing module-local capability | None proven by the frozen characterization and runtime tests |

Money remains checked integer minor units with pinned currency exponent;
cross-currency aggregation, float authority and implicit conversion remain
forbidden. Manual provider is a supported provider profile and deliberately has
no webhook implementation; the canonical webhook ingress rejects unsupported
provider use before reading a body.

## Migration and storage proof

- 416 Billing SQL documents were inventoried: 208 fresh plus 208 V1-adopted
  documents.
- Every fresh document has one V1-adopted peer with equal migration content;
  the module test fixes the expected pair count at 208.
- Billing versions span `20260717010000` through `20260802232805`.
- The controlled reopen added exactly six paired versions:
  `20260802232800` storage, `20260802232801` inspect,
  `20260802232802` feature resolution, `20260802232803` mutation record,
  `20260802232804` reconciliation inspect/storage, and `20260802232805`
  reconciliation record.
- A read-only repository-wide scan found zero Billing version/scope collisions
  with another migration owner and zero duplicate Billing origins.
- Prior Billing generations are present in the global migration stream. The six
  controlled-reopen versions and their exact storage/routine claims remain a
  locked, single-owner integration residue; they were deliberately not added to
  shared inventories during module work.

## Verification evidence

All commands were run from the `backend-v2` repository root.

| Check | Result |
| --- | --- |
| `go test ./core/internal/modules/billing/... -count=1` | PASS, all Billing packages |
| `go vet ./core/internal/modules/billing/...` | PASS |
| `go test ./tools/internal/characterization -run BillingV1 -count=1` | PASS |
| Canonical/license route constructor tests | PASS |
| V1 exact-route/binding/order tests | PASS |
| IAM exact Billing permission tests | PASS, 54 accepted pairs |
| Full Billing PG18 integration package | PASS, 355.339 s |
| PG18 provider subset | PASS, 25.783 s |
| PG18 subscription cancellation race | PASS, 12.211 s; exactly once |
| Metered-resource finite boundary | PASS on PG18; 2 concurrent creates produced 1 admission, 2 receipts, 1 projection, gauge 1 |
| Metered-resource owner rollback | PASS on PG18 fresh and V1-adopted; receipt, projection, usage decision, entry and gauge all rolled back |
| Metered-resource runtime ACL | PASS on PG18 fresh and V1-adopted; 0 direct table DML, all 5 typed routines executable |
| Metered-resource migration application | PASS on PostgreSQL 18.4, fresh and V1-adopted |
| Analytics consumer package after public-contract change | PASS, all Analytics packages |
| Full fileguard filtered to `core/internal/modules/billing/` | zero Billing findings; global exit remains non-zero outside Billing |
| Full archguard filtered to `core/internal/modules/billing/` | zero local dependency or SQL-shape findings; only locked global registration/claim residue remains |

Running fileguard with `-root core/internal/modules/billing` alone reports three
false provenance misses because that scoped invocation cannot see the
repository-level `contracts/generated.yaml`. The correct full-repository run
recognizes all three exact hashes and reports zero Billing-local findings.
Global fileguard/archguard currently exit non-zero. Billing's remaining
archguard output is limited to shared operation registration, storage-directory
claims and exact platform table/routine claims. Those files are locked until
the one-time integration pass and are not a reason to reopen this module.

## One-time integration-owner work after all freezes

1. Construct `billing.NewMeteredResourceEntitlements(...)` once in final
   composition and inject the same instance into the Analytics query,
   dashboard and publication applications. Do not add a second counter or
   entitlement resolver in Analytics/API.
2. Bind an authoritative Analytics/API owner-evidence verifier for
   `billing.NewMeteredResourceReconciliation(...)`; owner evidence must be read
   in the same live `TenantScope` before adoption or repair.
3. Register the final cross-owner operation authority/participants and the
   Billing operation-bearing source files once in the central operation
   manifest.
4. Add the exact directory claim for
   `core/internal/modules/billing/adapters/postgres/resourceentitlement/sql/platform`
   and exact claims for tables
   `platform.billing_metered_resource_receipts`,
   `platform.billing_metered_resources`,
   `platform.billing_metered_resource_reconcile_receipts` and routines
   `platform.billing_metered_resource_inspect_v1`,
   `platform.billing_metered_resource_feature_v1`,
   `platform.billing_metered_resource_record_v1`,
   `platform.billing_metered_resource_reconcile_inspect_v1`,
   `platform.billing_metered_resource_reconcile_record_v1` in the single
   storage-owner normalization pass.
5. Add versions `20260802232800` through `20260802232805` to the central
   migration inventory/frontier and normalize global expected counts once.
6. The central `contracts/characterization/v1-capability-parity.v1.json` has no
   Billing entry. During the single final parity normalization, explicitly
   reference or classify the already-frozen ADR-0075 Billing domain suite. Do
   not manufacture a feature count from routes, observers or fixtures.
7. `docs/characterization/billing-v1.md` records obsolete Merkle and semantic
   hashes. The current manifest hashes are
   `8876291015fb2f49b8e8b85f9bfe092c37dfe89b522158396fbafb53810da8f0`
   and
   `9cd7bcf7df954f039c7633fb142ec8f1c88d4a76c27c7cad7ae613c5e393eaf1`.
   Reconcile that global documentation once, without regenerating Billing.
8. After every module handoff exists, normalize permission/operation manifests,
   migration generation order and global expected-count gates once under the
   single integration owner.
9. Run targeted frozen-module suites, then global fileguard/archguard/parity,
   apply migrations to a clean real PostgreSQL instance, and start the complete
   local stack.
10. Through Gateway port 8000, live-test login, tenant selection, Billing catalog
   and subscription reads/writes, usage enforcement, platform reports/license,
   provider processing and the V1 retirement/compatibility paths. Unit and PG18
   package success alone is not product runtime acceptance.

Until step 1 begins under the integration owner, global manifests and migration
inventories remain locked. A later endpoint elsewhere does not invalidate this
Billing freeze unless it demonstrably changes a Billing-owned contract.

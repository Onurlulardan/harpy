# Tenancy aggregate freeze handoff

Status: **FROZEN** (module-local), 2026-08-02.

Reopen this aggregate only for a reproducible Tenancy-owned defect. Global
composition, manifests, migration ordering and parity normalization belong to
the single final integration owner.

## Closed capability gap

IAM already catalogued `platform.tenants.read`, `update`, `suspend`,
`reactivate`, `offboard` and `purge`, but Tenancy had no application, HTTP or
storage implementation behind those capabilities. The permissions were ghost
capabilities. This freeze adds the missing owner slice without changing frozen
IAM, Catalog or Tenant Settings code.

The canonical SaaS control-plane surface is now:

- `GET /api/platform/tenants`
- `GET|PUT /api/platform/tenants/{tenant_id}`
- `POST /api/platform/tenants/{tenant_id}/{suspend|reactivate|offboard|purge}`
- `GET /api/platform/tenant-lifecycle-operations/{operation_id}`

Every command reads the persisted deployment policy first, admits only SaaS,
requires the exact `saas_control` primary target and rejects that control tenant
as a customer target. IAM evidence is checked only after that boundary.

## Legacy disposition

The legacy Gateway had nine tenant-admin routes. All have an evidence-backed
disposition:

| Legacy capability | V2 disposition |
| --- | --- |
| list/detail | exact compatibility edge over bounded reads |
| create | existing durable customer provisioning command |
| update name | revisioned and idempotent compatibility edge |
| update status | `426`; use explicit suspend/reactivate action |
| suspend/reactivate | revisioned immutable receipt |
| delete | async offboard; synchronous schema DROP is forbidden |
| stats | `426`; final owner-projection composite, no cross-owner SQL |
| switch tenant | `410`; ADR-0117 forbids operator impersonation |

Anonymous `POST /api/auth/register-tenant` remains the existing `410`
retirement. Legacy `owner_id` and raw `settings` are not copied into the
platform lifecycle DTO: authority belongs to IAM and settings remain the
Tenant Settings contract. The control response exposes separate
`lifecycle_revision` and `settings_revision` values.

## Storage and safety contract

Paired immutable migrations `20260802232300..20260802232305` add:

- `lifecycle_revision`, `offboarded_at` and `purged_at` lifecycle state;
- immutable update/suspend/reactivate idempotency receipts;
- leased async offboard/purge operations with terminal mutation guards;
- narrow list/get/mutate/accept/read/claim/complete/fail owner routines;
- no runtime table privileges and exact routine-only runtime grants.

Offboard immediately closes tenant runtime admission through the existing
tenant status/deleted resolver and never drops a schema. Purge is separate and
cannot be accepted from caller booleans: a trusted readiness port must provide
retention, recovery/backup and contributor receipts plus a bounded evidence
digest. Purge completes only after an external destruction executor returns an
exact receipt. Ambiguous purge execution records `UNKNOWN`, not false success.

Fresh and adopted origins contain 127 platform migrations each. All six new
pairs are byte-identical. Their fresh SHA-256 values are:

- `32300`: `1F784DFABFF47014FE25A9198DDBC41240AA7BE55A171CA616C7C6A908EFD3BD`
- `32301`: `1CDF7C75B27B6CD6D6DA930A5691FDAD9EAFC6CDD1D046933024B4B9C956ACE5`
- `32302`: `0234721C21024EB23012B44AA14DD7382BA58B8DCB7CFB362EB59B74D04A9858`
- `32303`: `4F5E3143524258618D57F500124FBB11ADA3A70520D9E41A17F812E7420990D8`
- `32304`: `5E29CECDB5B72B4C4755674E34A6782C2019FED68F4EE08AB38485C11322EA46`
- `32305`: `C40CD03D2AB0D8C43AE3950493555AC74CAA8F4B814785DB7DB9EB18986E3045`

## Verification evidence

- `go test ./core/internal/modules/tenancy/... -count=1`: PASS.
- `go vet ./core/internal/modules/tenancy/...`: PASS.
- scoped `fileguard`: PASS.
- targeted Tenancy archguard tests: PASS.
- PostgreSQL 18.4 full migration application, fresh and adopted: PASS.
- PostgreSQL 18.4 lifecycle behavior, fresh and adopted: PASS, including
  bounded/control-excluding reads, replay/conflict, revision conflicts,
  suspend/reactivate, async offboard, proofed purge, narrow ACL and terminal
  receipt immutability.

## Locked single-owner integration residue

Do not solve these by reopening Tenancy:

1. Compose `CustomerLifecycleHTTP`, the API store, generic frozen IAM
   `PlatformAccessRuntime` adapter and its routes.
2. Compose/schedule the lifecycle worker plus trusted purge-readiness and
   idempotent schema-destruction ports.
3. Add exact operation, authority, storage-directory and platform-routine
   claims. Current full archguard findings for this slice are precisely stale
   locked manifests, matching the already-frozen Tenant Settings residue.
4. Add the six paired migrations once to the global migration order/frontier.
5. Aggregate Usage, Billing, Forms and IAM owner projections for the legacy
   stats successor without cross-owner SQL.
6. Normalize global OpenAPI, permission/parity inventories and fixed route
   counts once, after every aggregate is frozen.
7. Prove on-premise denial, login/control-tenant access and the full lifecycle
   through Gateway during final live E2E.

No globally locked manifest, aggregate OpenAPI, composition, migration
registry/frontier, permission catalog or parity snapshot was edited during this
freeze.

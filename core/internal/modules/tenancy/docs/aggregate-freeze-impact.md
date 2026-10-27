# Tenancy aggregate freeze impact

Status: `IN_PROGRESS`. This file is module-local. It does not authorize edits to
the locked global operation, authority, OpenAPI, parity, permission or migration
inventories.

## Intent and authoritative boundaries

The Tenancy aggregate must close its own platform tenant lifecycle before it is
frozen. ADR-0116, ADR-0117, ADR-0119 and the accepted Phase 14 platform
capability matrix make Tenancy the owner of the SaaS customer registry and the
following exact platform capabilities:

- `platform.tenants.read`
- `platform.tenants.create`
- `platform.tenants.update`
- `platform.tenants.suspend`
- `platform.tenants.reactivate`
- `platform.tenants.offboard`
- `platform.tenants.purge`
- `platform.tenants.admin_setup.reissue`

Mode and target admission precede IAM authorization. These capabilities are
SaaS-only, the actor must remain in the immutable control tenant, and the
control tenant can never be a lifecycle target. A platform subject does not
gain customer-tenant authority or impersonation.

The already frozen Tenant Settings slice is excluded from implementation work.
Its three routes, migration and owner ports must not be reopened without a
reproducible defect.

## Legacy capability inventory

The V1 auth service exposes nine tenant administration routes beneath the
Gateway admin proxy:

| Legacy operation | Module-local disposition required before freeze |
| --- | --- |
| `GET /api/admin/tenants` | exact compatibility edge over bounded customer list |
| `POST /api/admin/tenants` | already redesigned as durable customer provisioning |
| `GET /api/admin/tenants/{id}` | exact compatibility edge over customer detail |
| `PUT /api/admin/tenants/{id}` | revisioned and idempotent customer update |
| `POST /api/admin/tenants/{id}/suspend` | durable receipt; customer target only |
| `POST /api/admin/tenants/{id}/reactivate` | durable receipt; customer target only |
| `DELETE /api/admin/tenants/{id}` | compatibility edge to async offboard; no schema DROP |
| `GET /api/admin/tenants/{id}/stats` | redesigned owner projections; no cross-owner SQL |
| `POST /api/admin/switch-tenant` | intentional retirement; ADR-0117 forbids impersonation |

Canonical V2 lifecycle routes must coexist with the admitted compatibility
edges without a second writer. Anonymous `POST /api/auth/register-tenant`
remains the existing `410` retirement.

## Owner and transaction impact

- Tenancy alone mutates `platform.tenants`, customer lifecycle operations and
  receipts through narrow owner routines.
- IAM supplies platform access evidence only. It does not write tenant state.
- Suspend/reactivate/offboard must advance the Tenancy revision and leave an
  immutable idempotency receipt.
- Offboard is asynchronous and must close login/runtime admission without
  deleting a schema.
- Purge is a separate asynchronous command. It cannot complete until exact
  retention, recovery/backup and contributor blocker evidence is verified.
  Caller-supplied booleans or raw SQL are not evidence.
- Legacy tenant statistics must use owner ports or the accepted Usage/Billing
  projections; Tenancy must not read Billing, IAM, Forms or Usage tables.
- No operation may target the `saas_control` binding or run in on-premise mode.

## Migration and verification boundary

If forward migrations are required, Tenancy owns the reserved
`20260802232300..` range. Every new `fresh` and `v1_adopted` pair must be
byte-identical. Historical origin-specific adoption migrations are immutable.

Freeze requires all of the following:

1. exact route and constructor tests, including typed-nil dependency denial;
2. mode-before-IAM, control-target, cross-tenant and capability matrices;
3. replay, digest conflict, revision conflict and concurrent mutation tests;
4. PostgreSQL 18 fresh/adopted ACL, receipt and two-tenant isolation evidence;
5. full Tenancy tests, vet, scoped fileguard and targeted archguard;
6. a final module-local handoff containing only exact single-owner integration
   residue.


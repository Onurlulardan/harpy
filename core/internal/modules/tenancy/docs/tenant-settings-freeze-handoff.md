# Tenant Settings freeze handoff

Status: `MODULE_LOCAL_FROZEN`; do not reopen without a reproducible targeted or
live-runtime defect. Shared authority, manifest and composition work below is
reserved for the single integration owner.

## Exact legacy surface

The V1 `auth-service` exposes exactly three Gateway routes and V2 now owns the
same method/path set in Tenancy:

- `GET /api/tenant/settings`
- `PUT /api/tenant/settings`
- `POST /api/tenant/settings/test-email`

The redesign keeps the legacy response fields, moves the tenant directory
record to its Tenancy owner, moves SMTP state/delivery to Notifications, and
replaces plaintext SMTP password reads with `password_configured` plus a
write-only secret. Updates use expected revision, request digest and an
idempotency receipt instead of V1's unfenced `Save`.

## Module-local closure

- The registry adapter binds logical owner `tenancy`, not the legacy
  `auth-service`/IAM caller. Owner binding is isolated in `bind.go`; other SQL
  adapter files depend only on the Tenancy application scope alias.
- Command authority is declared as exact `tenancy` +
  `iam_system_resources`, adding `notifications` only for SMTP mutations.
  Database participants are `tenancy` and, when needed, `notifications`; IAM is
  authority evidence, not the tenant-directory writer.
- Snapshot authority declares exact owners `billing`, `forms`, `iam`,
  `notifications`, and `tenancy`. Capability minting is an injected platform
  owner port; the Tenancy module does not mint it.
- Fresh and `v1_adopted` migration
  `20260801020000_create_tenancy_settings_runtime_v1` are byte-identical
  (SHA-256 `c0c990918ec6e035838364aff58cb933133081bbf6b697be92a6661efcac4322`)
  and the version occurs only in Tenancy. The new read routine is hardened
  PL/pgSQL without forbidden `STABLE`; runtime receives routine execution only.
- Module migration counts are 250 documents: 121 platform migrations per
  origin plus four tenant migrations per origin.

## Verification

- `go test ./core/internal/modules/tenancy/... -count=1` -- pass.
- `go vet ./core/internal/modules/tenancy/...` -- pass.
- `go run ./tools/fileguard -root core/internal/modules/tenancy` -- pass.
- `go test -tags=integration ./core/internal/modules/tenancy/adapters/postgres -run '^TestTenantSettingsRuntimeAgainstPostgreSQL184$' -count=1 -timeout 12m` -- pass against exact PostgreSQL 18.4 for both origins, including runtime ACL, replay/conflict, concurrent CAS, V1 revision fence, receipt immutability and two-tenant isolation.
- Filtered global archguard has no remaining module-layer, restricted-import,
  owner-bind, SQL-option or broad-permission defect in this slice.

## Single-owner integration debt

No item below authorizes a Tenancy module edit:

1. Register the `tenancy` tenant command-authority group in the shared
   transaction authority registry and test exact member/participant denial.
2. Provide the platform snapshot-admission implementation for the exact
   five-owner Tenant Settings projection. REST Query's ADR-0051 four-owner
   query limit remains separate and must not be silently widened.
3. Wire the concrete generation resolver, snapshot authority owner,
   coordinator, projection owner APIs and HTTP runtime in composition.
4. Register the four operation-bearing Go sources, two static SQL documents,
   three platform routines and the receipt table in the locked global
   operation/storage authority surfaces.
5. Add the paired migration to the locked release/checksum/frontier inventory.
6. After normalization, prove all three routes through Gateway with a real IAM
   decision, tenant selection, redacted GET, name update/replay, SMTP update and
   test-email enqueue.

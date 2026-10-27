# Support strict-freeze impact

## Scope

- Writable scope is only `core/internal/modules/support/**`.
- PlatformOps and every other module remain frozen and read-only.
- Shared composition, HTTP/permission/operation/storage manifests, migration
  inventory and parity files remain owned by the single integration pass.

## Current evidence

- The Support facade declares the 17 legacy contact, booking and bug-report
  routes, but the core API composition root does not construct or mount it.
- Targeted Support tests, vet, fileguard and PostgreSQL 18 fresh/adopted tests
  pass. Filtered archguard has no domain dependency, embedded SQL,
  adapter-direction or same-basename operation-test violation.
- The remaining filtered guard output is only absent shared operation/storage
  registration; those locked files belong to the final integration owner.
- The locked parity rule still classifies nine Support routes as missing even
  though the module now implements all 17 legacy route shapes.

## Applied changes

- Keep RFC 8785 canonicalization at the application/adapter boundary and remove
  non-standard-library dependencies from the domain package.
- Move runtime/adopter SQL statements to owned module-local SQL documents.
- Expose one module-owned runtime constructor which assembles PostgreSQL
  contact/bug drivers, resource operation services and all 17 HTTP routes.
- Split operation sources by contact, reporter, tenant, platform and adoption
  resource so every operation tuple and same-basename test is unique before the
  shared manifest is generated.

## Invariants and risk

- No route, permission string, table/function name, migration version or
  fresh/adopted migration body may change.
- No direct runtime table grant or caller-supplied SQL/identifier is introduced.
- Contact honeypot/rate/idempotency behavior, reporter redaction, tenant and
  platform permission scopes, optimistic revision checks and adoption evidence
  remain unchanged.
- Verification requires targeted unit/vet/fileguard/filtered-archguard plus real
  PostgreSQL 18.4 fresh and adopted migration/adoption execution.

## Rollback

The source refactor is mechanically reversible because persisted schemas and
migration documents are unchanged. A failing behavior or PostgreSQL gate blocks
freeze; no global registration is attempted as a workaround.

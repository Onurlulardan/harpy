# Audit strict-freeze impact

## Owner and scope

- Owner: `audit`.
- Public HTTP compatibility surface: the five authenticated legacy changelog
  `GET` routes only.
- Internal write surface: owner modules append a sealed fact inside an existing
  tenant command scope. Audit does not expose public ingestion.
- This pass may change only `core/internal/modules/audit/**`.

## Contracts and dependencies

- HTTP reads require the exact `changelog.read` permission and an exact
  operation ID for list, stats, get, entity history, or form-record history.
- The module constructor composes its own HTTP adapter, application, PostgreSQL
  binder, and authority evidence. Final gateway composition remains external.
- Domain values are standard-library-only. UUID conversion belongs at the
  application or adapter boundary.
- PostgreSQL adapters receive only the module access mode. Only `bind.go` may
  import and bind the opaque platform tenant scope.
- SQL statements are immutable embedded files under the owned PostgreSQL
  adapter leaf; Go source contains no PostgreSQL statements.

## Data, tenancy, and reliability

- Owned tenant objects use the `audit_*` prefix. Runtime roles receive routine
  execution only; direct table mutation and reads stay denied.
- Audit entries are immutable and append-idempotent by owner audit identity and
  request digest. New audit IDs are UUIDv7; adopted IAM identities may be UUIDv4.
- Confidential and secret field values and their digests are never serialized
  to HTTP. Only field name, classification, and presence flags are returned.
- Reads are bounded and ordered by `created_at DESC`; unknown query keys and
  alternate sort expressions are rejected.
- No queue or eventually consistent audit write is introduced. Business-owner
  wiring must call the sealed append boundary in the originating transaction.

## Legacy compatibility decision

- Exact: five route paths, read permission intent, list filters, pagination,
  details, entity history, form-record history, and aggregate statistics.
- Redesigned: raw `old_data`, `new_data`, request metadata, IP address, user
  agent, and arbitrary public `/ingest` are intentionally absent because the
  immutable projection is redacted and transaction-bound.
- Adopted: legacy `change_logs` rows are copied in bounded batches with source
  count, cursor, digest, and drift fences; raw old/new values are not copied.

## Rollout and verification

1. Repair only module-local dependency direction, identity, and SQL placement.
2. Prove endpoint, application, adapter, permission, constructor, and legacy
   mapping contracts with targeted tests.
3. Run Audit tests, vet, scoped fileguard, filtered archguard, and the exact
   PostgreSQL 18.4 fresh/adopted integration test.
4. Freeze the module and hand only operation, authority, storage, migration
   inventory, composition, and live-gateway registrations to the single global
   integration owner.

## Open decisions outside this module

- Global operation/authority fragments for the five reads.
- Global storage ownership claim for Audit migrations and static SQL.
- Global migration ordering/generation inventory for the seven paired versions.
- Composition of the HTTP facade, owner append boundary, and adoption executor.
- Gateway-level login, tenant, authorization, and live route verification.

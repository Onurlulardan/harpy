# Audit module freeze handoff

Status: module-local implementation frozen; global integration intentionally
pending with the single integration owner.

## Frozen boundary

- Owner: `audit`; tenant-scoped storage; `audit_*` relations and routines only.
- HTTP compatibility surface: five authenticated legacy changelog `GET` routes.
- Read authority: exact `changelog.read` permission plus one exact operation ID.
- Write surface: no public ingest route. Owner modules append sealed audit facts
  inside their existing tenant command transaction.
- Runtime storage access: routine execution only. Runtime and adopter roles have
  no direct table read or mutation grant.
- Source boundary: domain uses standard-library identities; platform tenant
  scope binding remains isolated in the PostgreSQL binder and root runtime.
- SQL boundary: adapter statements are embedded from
  `adapters/postgres/sql/tenant`; SQL is not embedded as Go string literals.

The module must not be reopened unless a targeted Audit test or a global
integration failure proves a defect inside this boundary.

## Capability and parity decision

| Legacy capability | V2 decision | Evidence |
| --- | --- | --- |
| `GET /api/changelog` | Exact route, filters, bounded pagination and fixed `created_at DESC` ordering | HTTP contract and PostgreSQL tests |
| `GET /api/changelog/stats` | Exact route and aggregate intent | HTTP contract and PostgreSQL tests |
| `GET /api/changelog/{id}` | Exact route and canonical non-zero UUID validation | HTTP and domain tests |
| `GET /api/changelog/history/{entity_type}/{entity_id}` | Exact route and entity history filter | HTTP contract test |
| `GET /api/changelog/forms/{form_slug}/records/{record_id}/history` | Exact route and form-record history filter | HTTP contract test |
| Legacy read permission | Exact `changelog.read` intent on all five routes | HTTP/runtime/application tests |
| Public/internal legacy `/api/changelog/ingest` | Intentional redesign: not gateway-exposed and not accepted in V2 | Transaction-bound sealed append API |
| Raw `old_data`, `new_data`, request path, IP, user agent and metadata | Intentional redesign: omitted; confidential changes expose presence/classification only | Response and PG18 redaction tests |
| Legacy `change_logs` data | Redesigned equivalent: bounded adoption with OID/count/cursor/hash/drift fences | PG18 fresh/adopted suite |

One legacy parity defect was repaired during freeze: an empty requested page
previously reported `total=0` because the list routine returns the total only
with a row. The adapter now performs a bounded page-1 probe with identical
filters only when a deep page is empty, so the legacy total and total-page
semantics remain correct.

## Frozen operation contracts

All operations are tenant snapshot reads and require `changelog.read`:

- `changelog.list`
- `changelog.stats`
- `changelog.get`
- `changelog.entity_history`
- `changelog.form_record_history`

The module constructor composes the PostgreSQL binder and context authorizer.
The HTTP constructor proves a ready runtime, handler and exactly five routes.

## Frozen migration and storage contract

- Paired origins: `fresh` and `v1_adopted`.
- Versions: `20260801040000` through `20260801040006`, inclusive.
- Migration directory: `core/internal/modules/audit/migrations/tenant`.
- Static SQL directory:
  `core/internal/modules/audit/adapters/postgres/sql/tenant`.
- Runtime routines and signatures:
  - `audit_entry_append_v1(uuid,uuid,uuid,uuid,uuid,text,text,text,text,text,text,bigint,text,jsonb,bytea,timestamptz,timestamptz)`
  - `audit_entries_list_v1(text,text,text,text,text,uuid,uuid,uuid,timestamptz,timestamptz,text,integer,integer)`
  - `audit_entry_get_v1(uuid)`
  - `audit_stats_v1(timestamptz)`
- Adopter-only routine:
  `audit_v1_adopt_batch_v1(uuid,uuid,integer,timestamptz)`.
- Legacy table claim: `change_logs`, accessed only by the bounded adoption
  routine; retirement gate name must be assigned during the one-time global
  cutover normalization.

The exact migration generation is deliberately not guessed here. The global
integration owner must place this already-frozen version range once against the
then-current repository generation sequence.

## Verification evidence

The frozen source passed:

```text
go test ./core/internal/modules/audit/... -count=1
go vet ./core/internal/modules/audit/...
go run ./tools/fileguard -root core/internal/modules/audit
go test -tags=integration ./core/internal/modules/audit/adapters/postgres/integrationtest -run TestAuditFreshAndAdoptedRoutinesAgainstPostgreSQL184 -count=1 -v
```

The pinned PostgreSQL 18.4 integration suite passed for both migration origins
and covers append replay/conflict, immutable storage, runtime/adopter direct
table denial, bounded adoption, restart, drift fencing and redaction.

The filtered architecture gate has zero remaining module-local findings. Its
24 remaining Audit-path findings are locked global registrations: five
operation-source registrations and 19 storage-owner reflections (four static
SQL files, one migration provider and fourteen paired migration files).

## Single-owner global integration claims

These are not module defects and were intentionally not edited during freeze.
The integration owner must perform them once after all modules are frozen:

1. Add Audit operation fragments and include them once in
   `contracts/operations.yaml` for the five operation-bearing sources:
   - `adapters/http/authorize.go`
   - `adapters/http/read.go`
   - `adapters/postgres/read.go`
   - `application/context_authorizer.go`
   - `application/read.go`
2. Add the five authorities once under `contracts/authority/` and include the
   fragment in `contracts/authority.yaml`; scope source is `actor_tenant`, owner
   and participant are `audit`, and required group/domain is `audit`.
3. Add one Audit storage-owner fragment under `contracts/storage/owners/` and
   include it in `contracts/storage/owners.yaml`. It must claim the tenant
   migration directory, static SQL directory, four runtime routine grants, the
   adopter-only routine grant and the `change_logs` legacy-table adapter.
4. Add one Audit migration-stream fragment for versions
   `20260801040000..20260801040006` and include it in
   `contracts/storage/migration-streams.yaml`, using the generation selected by
   the integration owner after all frozen streams are known.
5. Add the Audit migration provider once to
   `core/internal/composition/migrator/migration_documents.go` in the normalized
   global migration order.
6. Add the Audit HTTP facade and IAM verifier bridge once to API composition.
   IAM must accept only the five exact `changelog.read` operation pairs.
7. Replace the global parity rule `audit-changelog-missing` in
   `contracts/characterization/v1-capability-parity.v1.json` with the evidenced
   exact/redesigned decision; do not change the legacy expected route count of
   five and do not count the non-gateway ingest handler as a sixth route.
8. Connect each owner mutation to the sealed append boundary and prove business
   rollback also rolls back its audit fact. Compose the adoption executor under
   the leased adopter authority.

After those one-time edits, run targeted IAM/API/migrator tests, the complete
architecture and parity gates, apply migrations to the real local stack, and
exercise login, tenant selection and all five routes through gateway port 8000.
Unit or PG18 module tests alone do not prove live gateway readiness.

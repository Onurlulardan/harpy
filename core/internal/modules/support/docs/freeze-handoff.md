# Support strict-freeze handoff

Date: 2026-08-02

## Verdict

**Module-local status: FROZEN. Runtime status: NOT REACHABLE.**

Support owns a complete, tested implementation of the 17 legacy contact,
booking and bug-report route shapes. The core composition root does not yet
construct or mount `support.NewRuntimeHTTPHandler`, and the shared HTTP,
authority, operation, storage, migration-stream and parity contracts do not yet
admit Support. Those files remain locked to the single final integration owner.

Do not reopen this module unless a failing Support test, PostgreSQL execution or
live Gateway request demonstrates a module-owned defect.

## Exact HTTP parity and authority

| Method and route | Scope | Permission | Operation |
| --- | --- | --- | --- |
| `POST /api/contact` | public | admission/rate policy | `support.contact.submit` |
| `GET /api/public/booking-url` | public | none | immutable booking read |
| `GET /api/admin/contact-messages` | platform | `platform.contact_messages.read` | `support.contact.list` |
| `GET /api/admin/contact-messages/stats` | platform | `platform.contact_messages.read` | `support.contact.stats` |
| `GET /api/admin/contact-messages/{id}` | platform | `platform.contact_messages.read` | `support.contact.read` |
| `PUT /api/admin/contact-messages/{id}` | platform | `platform.contact_messages.manage` | `support.contact.update` |
| `POST /api/bug-reports` | reporter | `bug_reports.create` | `support.bug.create` |
| `GET /api/bug-reports/mine` | reporter | `bug_reports.read_own` | `support.bug.mine` |
| `GET /api/bug-reports/{id}` | reporter | `bug_reports.read_own` | `support.bug.read` |
| `GET /api/bug-reports/admin` | tenant | `bug_reports.read` | `support.bug.tenant.list` |
| `GET /api/bug-reports/admin/stats` | tenant | `bug_reports.read` | `support.bug.tenant.stats` |
| `GET /api/bug-reports/admin/{id}` | tenant | `bug_reports.read` | `support.bug.tenant.read` |
| `PUT /api/bug-reports/admin/{id}` | tenant | `bug_reports.manage` | `support.bug.tenant.update` |
| `GET /api/admin/bug-reports` | platform | `platform.bug_reports.read` | `support.bug.platform.list` |
| `GET /api/admin/bug-reports/stats` | platform | `platform.bug_reports.read` | `support.bug.platform.stats` |
| `GET /api/admin/bug-reports/{id}` | platform | `platform.bug_reports.read` | `support.bug.platform.read` |
| `PUT /api/admin/bug-reports/{id}` | platform | `platform.bug_reports.manage` | `support.bug.platform.update` |

The route count, method/path set and all 15 protected scope/permission/operation
tuples have exact contract tests. The same 17 method/path shapes exist in the V1
`auth-service` registration.

## Parity classification

- **Exact:** all 17 edge method/path shapes and the user-visible contact,
  booking, reporter, tenant-admin and platform-admin outcomes.
- **Redesigned equivalent:** V1 raw/mixed bodies become a stable envelope with
  correlation ID and `no-store`; one canonical platform bug table replaces
  cross-tenant schema iteration; UUIDv7, command receipts and revisions replace
  UUIDv4/last-write-wins behavior.
- **Intentional difference:** contact detail reads no longer mutate `new` to
  `read`; writes require `Idempotency-Key`, request digest and, for updates,
  `expected_revision`; reporter reads are actor-and-tenant pinned and redact
  admin notes.
- **Intentional difference:** contact IP/user-agent data is adoption evidence,
  not canonical-row data. Screenshot/attachment upload is not accepted until a
  Documents-owned contract exists.
- **Intentional difference:** query parsing is strict and uses bounded legacy
  page/page-size OFFSET with maximum offset 5000. It is not keyset pagination;
  the locked `docs/phase14-support-impact.md` statement must be corrected once
  during final integration.

## IAM root cause

- `bug_reports.create` and `bug_reports.read_own` are absent from IAM catalogs,
  tenant resource/action publication, permission migrations and role grants.
  Reporter endpoints will therefore fail closed when composed.
- `bug_reports.read` and `bug_reports.manage` currently exist only as tenant
  session-capability catalog strings; no matching tenant permission/resource
  migration or role grant was found. These are ghost capabilities today.
- `platform.bug_reports.read|manage` and
  `platform.contact_messages.read|manage` already exist in the platform IAM
  catalog and paired platform migrations.
- Final IAM integration must publish and grant the four tenant permissions with
  the intended default roles, then prove reporter-own and tenant-admin denial
  boundaries through the real verifier. Support must not weaken checks to work
  around missing IAM publication.

## Operation registration handoff

Archguard sees 34 unregistered operation-bearing sources: 18 application and 16
PostgreSQL adapter sources. Every source has a same-basename test and there are
no duplicate resource/operation tuples.

- Application contact: `application/contact/{submit,list,read,get,update}.go`.
- Application reporter bug: `application/reporterbug/{create,list,get}.go`.
- Application tenant bug: `application/tenantbug/{list,read,get,update}.go`.
- Application platform bug: `application/platformbug/{list,read,get,update}.go`.
- Application adoption: `application/adoption/{start,execute}.go`; register
  authorities `support.adoption.start` and `support.adoption.execute` with
  credential-target scope.
- PostgreSQL contact/reporterbug/tenantbug/platformbug have the same resource and
  verb sets, except adoption. Map each of those 16 adapter entries to its exact
  application source with `application_operation`.

HTTP authorities use the 16 operation IDs in `ports/operations.go`. Adoption is
an internal durable job and has no HTTP route.

## Storage and migration handoff

- Provider: `support.MigrationDocuments()` embeds 13 platform `fresh` and 13
  platform `v1_adopted` documents, versions `20260801019000` through
  `20260801019012`.
- All 13 fresh/adopted pairs are byte-identical; no Support version collides
  with a non-Support migration.
- Archguard reports 43 uncovered Support SQL files: 10 runtime adapter SQL, 7
  adoption adapter SQL, 13 fresh migrations and 13 adopted migrations. Register
  exact directory/table/routine ownership once; do not move or rewrite SQL.
- Tables: `platform.support_contact_messages`, `support_bug_reports`,
  `support_command_receipts`, `support_public_rate_leases`,
  `support_adoption_runs`, `support_adoption_mappings` and
  `support_adoption_page_receipts`.
- Runtime routines are the `platform.support_*_v1` contact and bug command/read
  routines. Adoption routines cover require/start/claim/apply/advance/reconcile/
  complete plus run/page/contact/bug evidence documents and target-drift checks.
- Runtime and adopter roles have no direct table access. Routines are hardened
  `SECURITY DEFINER` with fixed `pg_catalog` search paths and profile-specific
  execute grants.
- Register the embedded provider and platform migration stream/order, then
  regenerate the release migration manifest once after every module is frozen.

## Verification evidence

Executed from the `backend-v2` repository root on 2026-08-02:

```text
go test -count=1 ./core/internal/modules/support/...                         PASS
go vet ./core/internal/modules/support/...                                  PASS
go run ./tools/fileguard -root core/internal/modules/support                fileguard: OK
go test -count=1 -p=1 -timeout=12m -tags=integration \
  ./core/internal/modules/support/adapters/postgres/integrationtest          PASS (PostgreSQL 18)
operation inventory                                                         18 app, 16 adapter, 0 missing tests
migration comparison                                                        13 pairs, 0 byte mismatch, 0 collision
```

PostgreSQL coverage includes fresh and V1-adopted migration execution,
idempotency replay/mismatch, public rate admission, optimistic revision,
reporter/tenant/platform isolation, runtime/adopter privilege denial,
two-tenant adoption, crash-resume, locator/checksum drift and reconciliation.

Full archguard is not globally green. Its Support-filtered output contains only
the 34 shared operation registrations, 43 shared storage coverage findings and
one migration-provider ownership finding listed above. It contains no Support
module-layer import, embedded SQL, duplicate operation or missing-test finding.

## Single integration-owner checklist

1. Construct `support.NewRuntimeHTTPHandler` and mount all 17 routes through the
   core Gateway-facing API, never a direct service listener.
2. Add the 17 HTTP/OpenAPI paths and route admission fixtures.
3. Complete the four tenant IAM permissions and role grants; retain the four
   existing platform permissions.
4. Register the 18 application operations/authorities and 16 adapter mappings.
5. Register Support SQL ownership and `support.MigrationDocuments()`.
6. Insert the Support platform migration stream in the final migration order and
   regenerate the global release manifest once.
7. Replace stale parity rule `support-contact-booking-bug-report-missing`: it
   counts only 9 routes and falsely says implementation/storage are absent.
   Admit all 17 with the classifications above.
8. Correct the stale keyset claim, then run global fileguard, archguard, parity,
   migration verification and Phase 14 gates once.
9. Prove live Gateway flows: contact, booking, reporter create/mine/detail,
   tenant admin list/stats/detail/update and platform list/stats/detail/update.

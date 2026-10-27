# IAM strict-freeze handoff

Status: **MODULE-LOCAL FROZEN on 2026-08-02**. IAM route, constructor,
authorization, persistence and paired-migration gates are green. Reopen this
module only when a reproducible targeted failure is proved to originate inside
IAM. Global manifests, composition, migration ordering, parity snapshots and
generated aggregates remain locked for the single final-integration owner.

This verdict does not claim that the complete local product is running. Clean
migration application and live Gateway journeys remain final integration work.

## Organization-owner boundary

The organization-owner slice is frozen and must not be reopened for adjacent
IAM or global integration work.

- Organization create requires a canonical, sorted, unique and non-empty owner
  set of 1..100 current active users from the same tenant.
- The HTTP boundary accepts canonical `owner_ids` and the singular legacy
  `owner_id` compatibility input. Supplying both, JSON `null`, an empty set or
  duplicate IDs fails closed. Update omits owner replacement unless one of
  those exact fields is present.
- Exact replacement promotes/reuses existing memberships. A removed active
  owner becomes an active member; a suspended owner remains suspended. Role
  assignment and relation references are preserved, and tombstoned relations
  are not revived.
- Affected users receive one deterministic authorization/revision bump. The
  owner graph, organization mutation and canonical receipt/replay complete in
  the same IAM `TenantScope` transaction.
- System-resource `owner_ids` is editable and create-required. It delegates to
  the same organization commands; there is no alternative generic Forms SQL
  writer or second public owner command.
- Runtime create/update use V3 routines. Delete uses forward V2: only a
  suspended organization with the expected revision and no live child, role,
  permission binding or non-revoked organization assignment can be deleted.
  Remaining memberships are terminal-tombstoned and affected user clocks are
  bumped in the same receipt transaction. Normal membership mutation still
  protects the last owner.
- The existing permission-operation identities remain
  `iam.identity.organization.create`, `.update` and `.delete`; this slice adds
  no new permission identity.

Frozen paired tenant versions are `20260802232000..20260802232005`:

1. owner replacement V1;
2. organization create V3;
3. organization update V3;
4. create/update V3 runtime ACL;
5. organization delete V2;
6. delete V2 runtime ACL.

## HTTP and constructor inventory

The route constructors expose 191 exact method/path entries. This is a route
inventory, not a count of independent product features.

| IAM owner surface | Entries |
| --- | ---: |
| Auth | 10 |
| Identity users/roles/organizations | 24 |
| Invitations | 6 |
| Lifecycle current / retired V1 | 25 / 25 |
| Permission requests | 13 |
| Portal users | 5 |
| Profile | 7 |
| RBAC | 26 |
| Post-restore | 2 |
| Internal MCP | 9 |
| System transfer canonical / V1 | 9 / 30 |

JWKS is a separate exact GET/HEAD handler with conditional read and mutation
method rejection. Tests pin method, path, handler presence, duplicate absence,
literal-before-parameter ordering, retired-route `410` behavior and constructor
rejection of nil or typed-nil dependencies. Full IAM tests compile and exercise
the root runtime graph, transport applications, PostgreSQL binders, workers and
recovery runtimes.

## Authorization and permission freeze

- Auth, Identity, RBAC, lifecycle, portal and transfer writes use owner-scoped
  commands and current IAM evidence; tenant or actor substitution fails before
  storage use.
- Public AI access pins exactly 54 permission-operation pairs. Billing pins 54,
  Catalog 23, Documents 32 and Workflow 20. Cross-pairs and wildcards fail.
- HTTP contract coverage proves non-empty IAM bridges for Analytics, Catalog,
  Documents, Forms lifecycle, Forms records and Workflow.
- Audit, Usage and Support now have IAM-native fail-closed HTTP verifier
  contracts. Audit pins five `changelog.read` operations, Usage pins eight
  tenant plus eight platform operations, and Support pins all 15 protected
  scope/permission/operation tuples without cross-scope aliases.
- Paired version `20260802234200` publishes protected exact `changelog.read`
  and `analytics.usage.read`. Changelog grants the three canonical default roles;
  usage only tenant admin. Aliases are only `Yönetici`, `Müdür`, `Kullanıcı`;
  protected tenant admin keeps `tenant_all` without an exact binding.
- Analytics dashboard export pins only `analytics.dashboard.export` with
  `.create`, `.get`, `.delete` and `.content`. Its tenant permission is
  published by a forward migration, and the Analytics worker factory uses the
  Analytics-owned evidence routine for pinned-actor reauthorization.
- Support tenant `bug_reports.create|read_own|read|manage` is now a real IAM
  resource/action/permission catalog. Reporter-own grants preserve V1's
  authenticated-user boundary; tenant read/manage grants are limited to the
  protected or named tenant-admin roles. Platform Support permissions continue
  to use the existing platform capability catalog.
- MCP availability and invocation re-evaluate the current actor and every
  binding. Only the explicitly frozen IAM, Notifications inbox and AI
  Observability aliases are accepted; no general rewrite exists.
- Portal access requires the portal authority and binding. Operator-only paths
  reject portal tokens. Post-restore admits only its two exact recovery paths.
- AI platform-owner persistence now hides the owner binder behind an IAM-local
  interface; no exported binder capability or direct non-binder PostgreSQL
  platform import remains.

## Paired migration proof

- Total documents: 1,504.
- Platform: 106 fresh plus 106 `v1_adopted`.
- Tenant: 646 fresh plus 646 `v1_adopted`.
- Version span: `20260716010000..20260802234200`.
- Every logical migration has exactly one peer and byte-identical SHA-256
  content across origins.
- A repository-wide read-only scan found zero IAM version collision with any
  other module and zero malformed or duplicate IAM origin pair.

Historical migration documents were not rewritten during strict freeze.

## Verification evidence

All commands were run from the `backend-v2` repository root after the final
module-local fixes.

| Check | Result |
| --- | --- |
| `go test ./core/internal/modules/iam/... -count=1` | PASS |
| `go vet ./core/internal/modules/iam/...` | PASS |
| `go run ./tools/fileguard -root core/internal/modules/iam` | PASS |
| `go test ./tools/archguard -run 'IAM|OwnerBind|RoutineIdentity' -count=1` | PASS |
| migration pair/hash/collision read-only scan | PASS; 0 structural, byte or collision defects |
| PG18 tenant observability authority fresh/adopted | PASS; 2 exact permissions, 8 evidenced grants, no unrelated/protected exact grant, authorization clock bump and protected tenant-all coverage |
| PG18 consumer authority fresh/adopted | PASS; Analytics export and Support permission/grant boundaries |
| PG18 auth, authorization, identity and RBAC target set | PASS; 117.170 s |
| PG18 lifecycle fresh/adopted | PASS; 42.582 s |
| PG18 permission invitation, portal user, REST Query and system transfer | PASS; 84.710 s |
| PG18 MCP OAuth internal HTTP and presented-access revocation | PASS; 404.891 s |

The widened PG18 run found and closed one real test-boundary defect: lifecycle
sessions used a fixed `retain_until` that eventually fell behind the required
revocation-plus-30-days window. The fixture now derives all session timestamps
from one PostgreSQL statement clock, so both origins remain time-independent.

The full architecture gate is still globally red. A filtered rerun produced no
finding for the new Audit/Usage/Support verifiers, Analytics export/worker
contracts or paired `20260802234000..20260802234200` migrations. Existing
locked operation, storage, migration and immutable-parser registrations remain
single-owner integration work.

## Exact single-owner final-integration residue

These items must be completed once without reopening IAM:

1. Publish the changed organization `owner_ids` field contract and catalog
   digest once in `system_resource_layout`; keep it editable and
   create-required.
2. Register IAM tenant versions `20260802232000..20260802232005` once in the
   global migration stream/release inventory and normalized generation order.
   Register `20260802234000..20260802234002`, `20260802234100` and `20260802234200` once; do not rewrite paired files.
3. Add the 11 existing IAM operation-bearing sources once: AI platform-owner
   `bind.go`/`read.go`, identity bulk-delete `reserve.go`, MCP OAuth adapter
   `access.go`, `decide_authorization.go`, `platform.go`, `refresh.go`, and MCP
   OAuth application `credential_scope.go`, `decide_authorization.go`,
   `revoke.go`, `rotate_refresh.go`.
4. Add exact storage claims once for the two AI platform-owner SQL files, the
   bulk-delete reserve SQL file, nine MCP OAuth static SQL files, the MCP
   credential-target resolver routine, and the paired MCP OAuth platform/tenant
   tables, routines and declared Tenancy foreign keys.
5. Classify the six immutable parser findings for paired versions
   `20260801016000..20260801016002` in the global gate; do not mutate historical
   SQL to silence the parser.
6. Bind `IAMMCPReadOwner` and `IAMMCPSessionOwner` once in MCP composition.
   Publish only the twelve frozen reads plus `session.whoami` and
   `session.logout`; never recreate `session.connect.*` as callable tools.
7. Bind `AIAccessRuntime` and `AIAuthorizationRuntime` to the frozen AI HTTP and
   worker consumers, and adapt `AIPlatformAuthorityRuntime` only to the four
   Observability current-instance-owner operations. Do not duplicate or widen
   IAM's exact matrices in composition.
8. Normalize permission, operation, storage, migration and parity aggregates
   once after every module handoff exists, then run global gates, apply a clean
   migration set and start the real local stack.
9. Through Gateway port 8000, live-test login, tenant selection, organization
   owner create/replace/delete, RBAC denial, portal access, profile, lifecycle,
   transfer and MCP/OAuth journeys. Green module and PG18 tests alone are not
   live Gateway acceptance.
10. Compose IAM Audit, Usage and Support verifiers once. Version `20260802234200`
    owns the tenant permissions; do not duplicate or bypass them. Bind
    `NewAnalyticsWorkerAuthorizationRuntime` for dashboard export refresh.
11. Keep portal unlink blocked until a Forms+IAM orchestration contract can
    atomically clear the Forms-owned subject link and mutate IAM access state.
    IAM alone cannot reproduce the V1 transaction safely.

No global manifest, composition, migration inventory, parity snapshot or
generated aggregate was edited by this freeze pass.

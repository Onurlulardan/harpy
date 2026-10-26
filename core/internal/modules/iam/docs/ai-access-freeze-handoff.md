# IAM AI authority freeze handoff

Status: **FROZEN** after the 2026-08-02 exact Usage-producer controlled reopen,
full IAM unit, vet and module-root fileguard closure.
Global composition, route/permission manifests, migration registries, parity
inventories and gate expectations remain single-owner integration work.

Controlled reopen on 2026-08-01 closed reproduced internal-usage, Plan-stage
and Memory producer/retrieval dead ends.
The public 54-operation matrix remains unchanged and still rejects every
internal operation. A separate non-HTTP runtime recognizes the seven internal
`ai.usage.*` operations plus `ai.plan.stage`, `ai.memory.retrieve` and
`ai.memory.record` from an exact valid public AI
permission/operation pair, requires the internal operation to match the active
AI-owned `TenantScope`, and re-resolves the actor at the captured authorization
version. Plan staging and retrieval accept only `ai.dispatch.stream`; recording
accepts only `ai.plan.execute` or `ai.plan.execute.stream`. Pair, source, scope
or version drift fails before IAM evidence is read. Six Usage operations have
admitted producer pairs; conservative reconciliation is recognized but has no
fabricated public source and fails closed. The public matrix was not expanded
by this additive owner contract.

The 2026-08-02 reopen replaced the former single-permission Usage derivation
with an exact permission-to-source-operation matrix. `ai.chat.use +
ai.dispatch.stream` admits only Usage effect lifecycle operations;
`ai.chat.use + ai.dispatch.estimate` admits only Usage budget;
`ai.workspace.manage + ai.workspace.rag.search.test` admits budget and effect
lifecycle operations for trusted rerank work. Exact `ai.chat.use` sources
`ai.plan.semantic.edit`, `ai.workspace.record.summary`, and
`ai.memory.retrieve` admit budget and effect lifecycle operations. Memory
retrieve is an existing internal operation whose durable job carries the
persisted original permission/source pair; admitting it here does not publish
an HTTP operation. Permission/operation cross-pairs and RAG reindex fail before
pinned IAM evidence is read. Conservative reconciliation has no fabricated
source and remains fail-closed. The public 54-operation matrix did not change.

## Frozen boundary

- `core/internal/modules/iam/ai_access_*.go`
- `core/internal/modules/iam/ai_authorization_runtime*.go`
- `core/internal/modules/iam/adapters/postgres/authorization/**` additions for
  the AI-owned evidence reader

No migration was added or renumbered. The existing read-only
`iam_authorization_evidence_read_v3` routine is reused.

A later controlled migration-only reopen did not alter this AI authority
contract. It moved the pre-existing IAM MCP OAuth platform sequence from
`20260801020000..003` to `20260801022000..003` and the tenant sequence from
`20260801021000..008` to `20260801022100..008`, eliminating collisions with
Tenancy and Notifications. Fresh/adopted pairs and SQL content are unchanged.

## Public tenant HTTP contract

`AIAccessRuntime` wraps the existing fail-closed `FormsRecordAccessRuntime`
security dependencies. It requires one valid Bearer JWT, exact optional tenant
headers, an operator-console token without a portal binding, and current IAM
authorization evidence containing an allowed permission ID and positive
authorization clock.

The exact public matrix contains 54 permission-operation pairs:

- config: 6
- interaction: 11
- knowledge: 6
- workspace tenant routes: 8
- plan lifecycle: 8
- execution: 4
- usage public reads: 2
- observability tenant routes: 6
- memory public administration: 3

Only these permission resources/actions are admitted:

- `ai.config.manage` -> `ai.config` / `manage`
- `ai.chat.use` -> `ai.chat` / `use`
- `ai.knowledge.read` -> `ai.knowledge` / `read`
- `ai.knowledge.manage` -> `ai.knowledge` / `manage`
- `ai.workspace.read` -> `ai.workspace` / `read`
- `ai.workspace.manage` -> `ai.workspace` / `manage`

Knowledge workers/diagnostics, plan staging, usage lifecycle/budget,
observability platform operations, workspace system-knowledge operations and
memory producer/embedding/curation operations are not in the public matrix.

## Durable application authorization

`AIAuthorizationRuntime.Authorize` accepts the AI-owned command scope, exact
principal and tenant, captured authorization version, permission and operation.
It requires exact scope actor/tenant ownership and uses
`NewPinnedActorRequest` plus `ResolvePinnedActor` to read the current active
actor token version while matching the captured authorization generation.
Allowed evidence must contain a permission ID and positive authorization clock.

`NewAIAuthorizer` seals the PostgreSQL evidence adapter to owner `ai`, command
mode, tenant-aware arguments and the V3 IAM evidence routine. Worker/internal
operations fail before the evidence adapter is called.

## Observability current-instance-owner authority

`AIPlatformAuthorityRuntime` is the IAM API for the four frozen Observability
platform operations: baseline freeze/list/get and global insight projection.
Its request has no tenant-ID or tenant-slug header fields. It accepts only the
exact `current_instance_owner` required authority; a generic analytics
capability is not an equivalent authority.

The verified operator JWT is rechecked against the current tenant session for
the exact session, client/access mode, token version, authorization version,
platform clock and security epoch. In the same IAM-owned command scope,
`platform.iam_platform_owner_context_lock_v1(tenant,user,expected_clock)` locks
and projects the current owner. The returned platform subject, environment,
deployment mode and platform clock must match current session evidence before
the decision sets `CurrentInstanceOwner=true`.

## Verification and integration handoff

Targeted tests cover the exact 54-pair matrix, wrong pairs, internal operations,
the exact Plan/Memory source allowlists,
Bearer/JWT failures, portal denial, tenant mismatch, IAM denial/unavailability,
scope actor/tenant mismatch, stale authorization generation, incomplete
evidence, all four current-owner operations, rejected generic authority,
session/token/auth/platform-clock/security-epoch drift, owner/environment drift,
SQL markers and adapter owner/mode.

Green commands on 2026-08-01:

```text
go test ./core/internal/modules/iam -run '^(TestAIAccess|TestAIAuthorization|TestAIPlatformAuthority)' -count=1
go test ./core/internal/modules/iam/adapters/postgres/authorization ./core/internal/modules/iam/adapters/postgres/aiplatformowner -count=1
go test ./core/internal/modules/iam/... -count=1
go vet ./core/internal/modules/iam/...
go run ./tools/fileguard -root core/internal/modules/iam
```

The complete IAM test suite, vet and module-root fileguard were repeated green
on 2026-08-02 after the exact producer-matrix reopen. No migration, SQL, public
permission or global integration file changed.

The first full IAM run exposed five exact owner-backed HTTP contract pairs that
had not reached their IAM bridges. The controlled IAM reopen admitted only
those proven pairs: Catalog demo get/removal-plan/install/remove and Forms
import resume. Catalog owner permission maps, Forms async access maps and global
path contracts already agreed on all five. The targeted bridge gate and the
subsequent full IAM run are green; no owner or global file was changed.

The single integration owner may now wire these two IAM runtimes into the
already frozen AI HTTP/application composition. The exact matrix in IAM is the
single permission-operation mapping source; composition must not duplicate or
broaden it. Observability composition needs only a narrow type adapter from its
`PlatformAuthority` port to this IAM API; it must not recreate the authority
logic. Global permission publication and route registration happen once after
every participating module is frozen.

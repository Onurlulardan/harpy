# Work Aggregate Impact

Status: **frozen** after owner-local closure on 2026-08-02. Reopen only for a
reproducible Work defect. Shared HTTP/OpenAPI, operation/permission/storage
manifests, migration frontier/inventory, parity snapshots and composition stay
locked for the single integration owner.

## Evidence boundary

- The real V1 gateway and workflow service expose 27 Work HTTP routes: four
  calendar routes, three static board-task reads, four board CRUD routes, five
  membership routes, four column routes, four task routes and three comment
  routes.
- The Work V2 compatibility fragment currently exposes 32 routes. Five are
  redesigned V2 compatibility conveniences rather than V1 route parity:
  `PUT /boards/{board_id}`, `GET|POST /board-tasks` and
  `GET /board-tasks/{task_id}`.
- The real V1 agent-tool registry owns ten Work capabilities. The Work-local
  MCP target set is exact, including prepare/confirm as two protocol
  capabilities rather than two HTTP features.

## Closed owner-local gaps

1. Canonical boards, tasks, comments, calendar events, inbox and unified
   calendar lists now use the shared REST Query resolver, authorization,
   compiler, PostgreSQL snapshot executor and signed continuation cursor.
   Compatibility reads remain on the bounded V1 adapter by design.
2. `/api/work/inbox` now unions assigned Work tasks with real
   `work_inbox_items`. `/api/work/calendar` is the ADR-0016 canonical union of
   Work deadlines, Work calendar events and `work_calendar_items`.
3. Every list has mandatory current-IAM principal visibility. Cross-owner rows
   additionally require current permission revision; stale rows fail closed.
   Missing/deactivated IAM targets and malformed principal revisions are denied.
4. Projection apply and repair now have a Work-owned runtime surface. Incoming
   visible principal sets are bounded and canonical, source authority is
   rechecked, and idempotency fingerprints bind the full visible payload.
5. Fresh and adopted migration `20260802232500` are byte-identical. It adds the
   query views and projection principal constraints, hides pre-evidence rows,
   and repairs omitted runtime column grants for task schedule/priority/color
   and projection visibility writes. V1 workflow projection adoption preserves
   principals through an exact active-run/stage match and fails closed when the
   match is absent or ambiguous.
6. The canonical route inventory is 68 routes after adding
   `GET /api/work/calendar`; the compatibility inventory remains 32 routes.
7. `FORM_FILL` and `FORM_REVIEW` now resolve through the Forms-owned work-task
   API. Explicit `form_data_id` resolves an exact record; a filter never
   fabricates record identity. Create/update registers or releases the Forms
   inbound read reference in the same Forms+Work transaction, and compatibility
   reads verify exact stable links before rendering aliases.
8. Migration `20260802232710` preserves V1 task recurrence and filter metadata
   in `work_task_compatibility_metadata` without adding recurrence columns to
   canonical V2 tasks. Runtime writes, Go adoption and SQL-only adoption use the
   same bounded domain contract.
9. Implicit-version V1 updates fingerprint the canonical V1 request intent,
   not the current task snapshot, row version or authority revisions. An exact
   key/body replay returns the original receipt after an unrelated task update;
   the same key with changed intent conflicts.
10. Work's principal-array SQL now uses PostgreSQL `ROWS FROM` semantics. The
    former schema-qualified multi-argument `unnest` call was not a real
    PostgreSQL function and failed on the runtime Reader path.

## Exact V1 compatibility dispositions

- V1's 27 gateway routes are represented. The five additional compatibility
  routes are redesigned conveniences, not five extra V1 features.
- V1 task recurrence fields only stored and returned metadata; no V1 scheduler
  consumed them. The compatibility adapter now round-trips that metadata while
  V2 executable recurrence remains a calendar event plus versioned recurrence
  rule. The legacy rule limit is the real `VARCHAR(50)` character limit.
  Invalid non-empty `recurrence_end` values are ignored, empty values clear and
  absent/JSON-null pointer fields preserve, matching the V1 handler.
- V1 create/update task DTOs did not accept `form_filter`, so compatibility
  writes intentionally ignore that field. Legacy adoption and reads preserve
  nullable or non-object JSONB values because the V1 database allowed them.
  The filter remains metadata only and never becomes a guessed record ID.
- `FORM_FILL`/`FORM_REVIEW`, `form_slug` and explicit `form_data_id` are closed
  owner-locally through the Forms API and stable resource links. Missing,
  malformed or revision-drifted Forms evidence fails closed.

## Integration residue (locked)

One integration owner must perform these once after all modules freeze:

1. Pass REST cursor keys and the Forms/Workflow source-owner authority into
   `PlatformHTTPRuntimeDependencies`; bind trusted event delivery to
   `ApplyProjection` and `RepairProjection`.
2. Add route 68 and `GET /api/work/calendar` to aggregate OpenAPI/readiness,
   and add migrations `20260802232500` and `20260802232710` to the migration
   inventory/frontier.
3. Normalize the global operation, permission and storage manifests without
   reopening Work.
4. Bind the completed Go Forms adoption resolver/reference registry in
   production composition. The production migrator still invokes SQL-only
   `work.v1_adoption.v1`; that SQL writes `FORM_LINK` rows without calling the
   Forms reference API. The integration owner must orchestrate the matching
   Forms reference registration/release or route the cutover through the Go
   adoption service before production cutover is complete.
5. Run global gates, apply both migration tracks, start the real local stack,
   then exercise login, tenant selection, Work authorization, inbox, unified
   calendar, projection freshness and all ten MCP capabilities through Gateway.

## Freeze evidence

- `go test ./core/internal/modules/work/... -count=1`: pass.
- `go vet ./core/internal/modules/work/...`: pass.
- PostgreSQL 18.4 fresh/adopted crash-replay isolation: pass.
- PostgreSQL 18.4 protected cutover atomic replay/evidence: pass.
- PostgreSQL 18.4 exact V1 `TIMESTAMP`/`VARCHAR(50)` adoption with nullable and
  object JSONB filters: pass.
- PostgreSQL 18.4 runtime role metadata ACL/upsert/load and real compatibility
  Reader round-trip: pass.
- PostgreSQL 18.4 REST Query visibility/revision/ACL test: pass.
- PostgreSQL 18.4 V1 workflow projection visibility exact/fail-closed test:
  pass.
- Fresh/adopted migration SHA-256:
  `358DC679B927B9AC298B58063B9ED4769AF02E2E0AE37359CD1B195BF53A9FE9`
  for `20260802232710`; the two tracks are byte-identical.

## Controlled worker-construction reopen

The nine-target Work execution surface spans the canonical list/get/write and
archive-preparation paths, so it needs the same real application graph as the
platform runtime but no HTTP handler. A typed facade will construct that graph
inside Work and expose only `MCPWorkExecutionOwner`. It rejects every partial
or typed-nil authority, scope, IAM principal, query, resource-reference and
async dependency. Direct-owner behavior and all HTTP/data/migration/event/job
contracts remain unchanged. Exact target inventory and full Work gates are the
acceptance evidence.

## Worker facade freeze evidence

Recorded on 2026-08-02:

- `NewMCPWorkExecutionProduction` reuses the real platform application graph,
  returns only the existing execution owner and constructs no handler;
- the owner exposes exactly nine executable targets, and a typed-nil principal
  authorizer is rejected before construction;
- `go test ./core/internal/modules/work/... -count=1` passed;
- `go vet ./core/internal/modules/work/...` passed;
- `go run ./tools/fileguard -root core/internal/modules/work` passed.

The facade is frozen. Global composition and archguard normalization remain the
single integration owner's work.

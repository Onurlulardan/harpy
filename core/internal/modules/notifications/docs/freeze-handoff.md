# Notifications Freeze Handoff

Status: Notifications-local implementation is frozen. Reopen only for a
reproducible Notifications defect. Shared MCP dispatch, composition and global
operation registration remain the single integration owner's work.

## Frozen MCP surface

The owner surface is exactly:

- `notifications.list` as a direct read and AI-executable read;
- `notifications.read` as a direct write and AI-executable idempotent command.

Both targets call the real `HTTPRuntime` and Notifications application paths.
They preserve the verified tenant and recipient, canonicalize arguments,
validate UUIDv7 identities, and fail closed on foreign results. Direct writes
require an MCP UUIDv7 idempotency key. AI execution binds target, input,
contract, grant, actor, tenant, plan, run, effect and fence evidence into the
receipt; reconciliation reads the persisted result instead of fabricating
success.

The legacy offset list is a bounded compatibility bridge over one stable V2
cursor snapshot. Unsupported provider-channel state, OR filters, alternate
sorts and unsupported read-state filters are rejected rather than guessed.
The response exposes `has_more` and `next_page`, not a fabricated exact total.

## IAM dependency

The previously open permission blocker is closed. IAM maps legacy MCP resource
`notifications` to canonical `notifications.inbox` only when the owner is
`notifications` and the action is `read` or `update`. Focused negative tests
reject delete, another owner and another resource without reaching the alias.

## Recorded gates

Run from `backend-v2`:

```text
go test ./core/internal/modules/notifications/... -count=1
go vet ./core/internal/modules/notifications/...
go run ./tools/fileguard -root core/internal/modules/notifications
go test -tags=integration ./core/internal/modules/notifications/... -count=1 -timeout 12m
go test ./core/internal/modules/iam -run 'TestMCPPermissionAuthorizerMapsOnlyFrozenNotificationsInboxAlias|TestMCPPermissionAuthorizerRejectsNotificationsAliasOutsideExactContract' -count=1
go run ./tools/archguard
```

Recorded result on 2026-08-02:

- all Notifications unit and contract tests passed;
- Notifications vet and scoped fileguard passed;
- all Notifications PostgreSQL 18 integration packages passed, including
  storage reliability, SMTP fresh/adopted, V1 adoption crash-resume, signed
  cutover and tenant-isolation/CAS paths;
- the exact IAM alias allow and denial tests passed;
- filtered archguard has no Notifications-local violation. Its sole
  Notifications finding is the global operation claim for
  `application/read_smtp_settings.go`.

## Single-owner integration debt

The integration owner must bind one `MCPInboxOwner` to the two direct targets,
one `MCPInboxExecutionOwner` to the same two AI targets, and register the one
SMTP-settings operation claim during final normalization. Final acceptance
must exercise both tools through the Gateway; module tests are not live route
proof.

## Controlled worker-construction reopen

The worker needs only the canonical inbox application behind the two frozen
execution targets. It must not construct the realtime ticket application or an
HTTP handler. The module-local facade therefore reuses the production
PostgreSQL binder, recipient resolver, scope authorizer, clock and IDs to build
the real inbox application, then returns the existing execution owner. Missing
or typed-nil generation, scope or async-catalog dependencies fail closed.
No endpoint, permission, data, migration, event or job semantics change.

## Worker facade freeze evidence

Recorded on 2026-08-02:

- `NewMCPInboxExecutionProduction` builds only the production inbox application
  and returns the existing execution owner; realtime and HTTP are excluded;
- the owner exposes exactly `notifications.list` and `notifications.read`, and
  a typed-nil scope coordinator is rejected before construction;
- `go test ./core/internal/modules/notifications/... -count=1` passed;
- `go vet ./core/internal/modules/notifications/...` passed;
- `go run ./tools/fileguard -root core/internal/modules/notifications` passed.

The facade is frozen. Global composition and archguard normalization remain the
single integration owner's work.

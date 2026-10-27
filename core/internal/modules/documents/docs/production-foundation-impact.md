# Documents production foundation impact

Status: controlled reopen complete; aggregate freeze evidence is in
`freeze-handoff.md`.

## Owner

- Documents remains the sole owner of application construction, encrypted
  artifact clients, clean-scan effects, Knowledge files and Documents MCP
  handlers.
- Composition receives only public application/owner interfaces. It does not
  import Documents adapters, application packages or nested storage ports.

## Affected contracts

- `ApplicationRuntime` is the transport-neutral Documents command/query
  runtime. `HTTPRuntime` and `NewHTTPRuntime` remain source-compatible aliases.
- `NewConfiguredAPIProduction` takes the existing API settings, async catalogs,
  generation resolver, tenant coordinator, Billing owner and build ID. It
  exposes `Application`, `KnowledgeFileOwner`, `MCPDirectOwner` and
  `MCPExecutionOwner`.
- `NewConfiguredWorkerProduction` takes the existing configured Documents job
  dependencies plus ClamAV address, generation resolver, application tenant
  coordinator and build ID. It exposes the Knowledge owner, MCP execution owner
  including its expiry reconciler, MCP execution owner and exactly six Documents
  job processors without constructing an HTTP handler.
- `NewConfiguredProduction` is the common configured root for callers that
  already hold typed artifact settings and a typed ClamAV address.
- Existing `NewConfiguredPlatformHTTPRuntime` and
  `NewConfiguredPlatformJobProcessors` constructors remain compatible wrappers.

## Data, tenant and permission impact

- No route, request/response, permission, operation, table, SQL, migration,
  event, job kind or retry policy changes.
- Existing `TenantScope`/`JobTenantScope`, authority generation, AI+Documents
  shared-owner and current IAM evidence checks remain unchanged.
- Typed nil, incomplete owner graphs and wrong processor counts fail before
  any owner is exposed.

## Resource and lifecycle ownership

- API opens one S3/keyring/cipher/ClamAV/Billing/async graph and shares it across
  HTTP application, Knowledge and MCP owners.
- Worker opens one target S3/keyring/cipher graph and shares it across jobs,
  Knowledge and MCP execution; the legacy source remains job-only.
- Scanner connections and returned document streams retain their existing
  per-operation close rules. These adapters expose no process-lifetime closer,
  so the immutable production bundles do not invent a second lifecycle API.

## Compatibility, rollout and rollback

- API can switch atomically from the compatibility wrapper to
  `NewConfiguredAPIProduction`; wire behavior is unchanged.
- Worker must add the existing canonical Documents ClamAV address to its typed
  profile, then replace the standalone job constructor with
  `NewConfiguredWorkerProduction`. This global config/composition edit remains
  locked to the single integration owner.
- Rollback to the compatibility wrappers is source-only and data-neutral.

## Required evidence

- Positive interface exposure and processor-copy tests.
- Nil, typed-nil, partial owner and partial processor negative tests.
- Full Documents tests, vet, scoped fileguard and targeted architecture tests.

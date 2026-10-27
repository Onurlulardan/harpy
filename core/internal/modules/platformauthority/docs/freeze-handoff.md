# PlatformAuthority strict-freeze handoff

Date: 2026-08-02

## Verdict

**Module-local status: FROZEN. Production composition status: REACHABLE.**

PlatformAuthority's protected-state, proposal, acknowledgement, signer,
registry, delivery, installation-head and recovery-evidence boundaries have no
known module-local blocker. Do not reopen this module unless a targeted test,
PostgreSQL execution or live runtime request proves that the defect belongs to
`core/internal/modules/platformauthority/**`.

This verdict does not mark the complete product or Phase 14 ready. Global
normalization, migration application and live Gateway journeys remain the final
single-owner integration pass.

## Frozen owner boundary

- Canonical protected authority state is environment- or tenant-scoped and is
  advanced only by a signed proposal plus an externally signed acknowledgement.
- Proposal signing and registry acknowledgement use purpose-separated trust;
  the application cannot acknowledge its own authority state.
- Air-gap, development journal and SaaS provider registry
  transports preserve the same generation and previous-head CAS semantics.
- Signer rotation fences stale signers and retains historical verification
  evidence. Replays are byte-identical; same-operation digest drift fails closed.
- Installation owner binding and protected-head pinning use narrow activation
  capabilities rather than normal user, JWT or platform permissions.
- Recovery evidence is read-only authority evidence consumed by the backup and
  restore profiles. PlatformOps remains the HTTP/recovery-operation owner.
- Delivery claim, retry and completion use durable leases and receipts; external
  effects do not silently advance current authority.

## Runtime and transport ownership

PlatformAuthority intentionally has no standalone HTTP listener and does not
own PlatformOps recovery routes. The production roots construct concrete module
adapters as follows:

- core API constructs on-premise recovery-anchor planning and SaaS customer
  authority planning with the PostgreSQL store and version-pinned KMS client;
- core worker constructs authority delivery and SaaS acknowledgement runtimes;
- core operator constructs initial SaaS authority, binds the installation owner,
  pins the acknowledged head and publishes development binding evidence;
- backup and restore migrator profiles consume the module's recovery-evidence
  constructor;
- PlatformOps owns `/api/platform/recovery-protection/**` and recovery operation
  HTTP surfaces and calls the PlatformAuthority planning ports.

Production compilation of API, worker and operator composition succeeded. No
test-only registry, signer or store is required by these constructors.

## Authority and tenancy decisions

- Protected authority is not a normal IAM permission and cannot be minted from
  HTTP, MCP, workflow, tenant headers or an authenticated-only session.
- Registry monotonicity proves latest external state; it does not select the
  Owner, Protected Admin or recovery anchor.
- Missing binding, wrong deployment mode, wrong environment/scope, stale head,
  wrong signer, absent activation capability and acknowledgement mismatch all
  fail closed.
- Environment and tenant authority heads remain distinct. SaaS tenant activation
  cannot write the environment head, and on-premise recovery keys are forbidden
  in the SaaS signer path.
- PostgreSQL routines are platform-scoped and activation capability calls are
  pinned to operation, environment, tenant, schema and operation kind.

## Migration and storage proof

- Provider: `platformauthority.MigrationDocuments()`.
- Scope: platform migrations only.
- Inventory: 22 fresh plus 22 V1-adopted SQL documents.
- Version span: `20260722040000` through `20260729100044`.
- Every fresh/adopted pair is byte-identical.
- A repository-wide read-only version scan found zero collision with another
  migration owner for all 22 versions.
- PostgreSQL 18.4 execution passed for both fresh and V1-adopted paths, including
  migrations, CAS replay/conflict, delivery, installation head capability,
  wrong-environment, wrong-kind and absent-capability denial.
- Filtered archguard reports no PlatformAuthority operation, storage, migration,
  SQL safety, import-direction or ownership finding.

## Parity classification

- **Infrastructure-only:** external protected-head CAS, proposal/ack trust,
  signer lifecycle and activation capabilities are V2 safety infrastructure,
  not separate replacements for user-facing V1 routes.
- **Redesigned equivalent:** V1-style privileged owner/recovery changes become
  durable protected operations that require current application authority,
  step-up at the PlatformOps edge and an external monotonic acknowledgement.
- **Intentional difference:** host deployment authority and
  recovery-kit custody never become Harpy IAM actors or Owners.
- **Intentional difference:** missing or stale external authority evidence keeps
  activation/recovery pending; it never falls back to database-only counters or
  trust-on-first-use.
- **Missing module-local capability:** none proven by the owner tests, production
  constructors and PostgreSQL integration suite.

## Verification evidence

Executed from the `backend-v2` repository root:

```text
go test ./core/internal/modules/platformauthority/...                     PASS
go vet ./core/internal/modules/platformauthority/...                      PASS
go run ./tools/fileguard -root core/internal/modules/platformauthority    PASS
go test -tags=integration \
  ./core/internal/modules/platformauthority/adapters/postgres/integrationtest \
  -count=1 -v                                                         PASS (15.166s)
fresh/adopted SHA-256 comparison                         22/22 identical
repository-wide migration version scan                    0 collisions
full archguard filtered to platformauthority              0 findings
go build ./core/internal/composition/api                              PASS
go build ./core/internal/composition/worker                           PASS
go build ./core/internal/composition/operator                         PASS
```

## Single integration-owner checklist

1. Keep this module frozen while the remaining IAM, MCP, AI owner and
   system-resource boundaries close.
2. During the one-time parity refresh, classify protected authority as the
   infrastructure/redesign decisions above; do not turn its routines,
   transports or migration count into fictitious missing product features.
3. Normalize global operation/authority, migration generation and release
   snapshots only after every module handoff exists.
4. Run frozen module suites, then global fileguard, archguard, parity and Phase
   14 gates once.
5. Apply the complete migration stream to clean and adopted databases, then
   start the real local stack without replacing data volumes.
6. Prove through the Gateway and real process profiles: recovery-protection
   denial/success, development authority delivery, installation owner/head
   activation, SaaS customer authority acknowledgement and stale-head denial.

Global gate drift or a new endpoint in another module is not evidence that this
module must be reopened.

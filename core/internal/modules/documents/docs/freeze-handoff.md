# Documents aggregate strict-freeze handoff

Status: **FROZEN - module-local gates green on 2026-08-02**.

The production-foundation controlled reopen was refrozen on 2026-08-02. It
changed no route, storage, migration or permission contract.

Documents is closed as one aggregate: HTTP, async jobs, encrypted storage,
purge, recovery, Forms lifecycle links/cleanup/archive, Knowledge files and MCP owners.
Do not reopen this module for global manifest, migration-order, OpenAPI, parity
or gate normalization. A reopen requires a reproducible Documents defect and a
new module-local handoff; otherwise the single integration owner completes the
pending global work below.

## Legacy capability classification

### HTTP

- All 20 legacy API Gateway route identities are present in the V2 router.
- Seventeen dispatch through V2 applications. Three legacy byte flows record
  their compatibility hit and return `426 CLIENT_UPGRADE_REQUIRED`:
  folder download uses async archives, and document/version creation uses the
  two-step upload contract.
- V2 has 32 routes total. The twelve additions are restore/purge/cancel,
  upload create/status/content/cancel, archive create/status/content and exact
  version-content download contracts.
- The six legacy service-internal temp/import-staging routes bypassed Gateway.
  Their capability is redesigned under the Forms import-source artifact owner,
  with encrypted object storage, durable source registration and cleanup. They
  are not missing Documents public routes.

### MCP and AI execution

- Direct MCP owns exactly thirteen legacy identities: seven reads, three
  side-effect-free prepares and three confirmed writes.
- AI execution owns the same seven reads and three confirmed writes. Prepare
  identities intentionally remain direct-MCP-only, for ten AI targets.
- Every target dispatches the real Documents runtime with actor, tenant, token
  version, authorization version/clock and exact permission/operation evidence.
- Intentional fail-closed differences remain: delete means 30-day trash rather
  than purge; list is folder-bound/cursor-based; root/null move and copy targets,
  legacy search/filter/sort/page-offset shapes and non-UUIDv7 identities fail.

### Forms linked-document archive

- Legacy Forms owns
  `GET /api/form-engine/dynamic/:slug/:id/documents/download`; it reads
  Documents rows and MinIO directly and builds one synchronous in-memory ZIP.
- `api/formrecordarchive` is the V2 Documents-owned command/query boundary.
  Creation accepts only a Forms-authorized definition/record/item set, refreshes
  current IAM `documents.item.download`, requires the exact active Documents
  form-record link set, then delegates to the canonical async archive command.
- Status is actor-owned, refreshes IAM again and binds the caller-supplied item
  set to the persisted archive request digest. Content remains the canonical
  `/api/document-archives/{id}/content` binary route.
- This is `redesigned_equivalent`: V2 returns a `202` archive handle and never
  restores synchronous ZIP, direct cross-owner SQL or Forms-owned MinIO access.

## Frozen module contracts

- HTTP and jobs use `TenantScope`/`JobTenantScope`; database mutation, owner
  receipt and job/outbox intent remain in the owning tenant transaction.
- Access evidence pins tenant and actor plus token/authorization revisions.
  Purge separately requires reason, recent step-up, cooling, reference epoch,
  inbound-reference and legal-hold checks, then durable deletion receipts.
- Bytes remain in private versioned object storage as authenticated `FCDOC1`
  ciphertext. Tenant/item/version identity, plaintext/ciphertext hashes, key ID,
  provider locator and clean-scan evidence are pinned; no DB byte payload or
  public/plaintext object contract is introduced.
- Recovery uses restricted backup/restore principals, DR holds that fence blob
  deletion, exact ciphertext inventory/copy receipts, locator remap evidence and
  deterministic transient terminalization.
- Forms cleanup cannot release links or create purge targets until the matching
  completed adoption receipt exists. Knowledge stage/adopt/release/open remains
  Documents-owned through the public owner boundary.
- Forms record archive idempotency is deterministically namespaced by definition
  and record. The existing archive transaction remains the only request, audit,
  job and receipt writer; the existing worker remains the only version pin,
  encrypted artifact, authorization-refresh, expiry and purge authority.

## Defects closed in this freeze

- API and worker composition previously had no configured Documents root that
  could share artifact clients across HTTP/application, Knowledge, MCP and job
  owners. Knowledge required eleven nested ports, while MCP execution accepted
  only `HTTPRuntime`. The new API/worker production bundles expose public owner
  interfaces, reject typed nil/partial graphs and keep old constructors as
  compatible wrappers.

- Forms cleanup previously joined adoption for the state transition only, so a
  missing receipt left cleanup `REQUESTED` while links were released and purge
  targets were created. `matched_links` now also requires adoption; unit and
  PostgreSQL 18 negative fixtures prove no premature effect.
- Recovery seal/release/fail and transient terminalization previously returned
  false/zero on exact crash replay. Additive `_v2` routines return the persisted
  exact result, reject substitution, retain restricted grants and leave applied
  `_v1` routines immutable. Both migration origins exercise the replay path.
- Forms had no public Documents owner API for the legacy linked-document archive,
  forcing any attempted parity implementation toward forbidden Documents SQL and
  object access. `NewFormRecordArchiveOwner` now constructs the real link/archive
  PostgreSQL adapters and existing durable async job port without a second state
  machine or storage path.

## Migration evidence

- Tenant tracks: 59 `fresh` plus 59 `v1_adopted`, from `20260718200100` through
  `20260802232400`; every same-version pair is byte-identical.
- Migration `20260802232400` is 200 lines and its paired SHA-256 is
  `33770BDF403BBEAEFAAA2D4318B6017FB5E360005ED59F485CF546123D07D7A2`.
- Active repo scan (supersession copies excluded): 4,266 SQL files and zero
  `(scope, track, version)` collisions.

## Acceptance evidence

```text
go test ./core/internal/modules/documents \
  -run '^Test(NewProduction|NewWorkerProduction|ConfiguredProductions)' -count=1
PASS

go test ./core/internal/modules/documents/... -count=1
PASS

go test ./core/internal/modules/documents/api/formrecordarchive \
  ./core/internal/modules/documents/application/formrecordarchive -count=1
PASS

go vet ./core/internal/modules/documents/...
PASS

go run ./tools/fileguard -root core/internal/modules/documents
fileguard: OK

go test ./tools/archguard ./tools/fileguard -run 'Documents|MCP' -count=1
PASS

go test ./tools/legacyparity -run 'MCP|Documents' -count=1
PASS

go test -tags=integration ./core/internal/modules/documents/adapters/postgres/integrationtest \
  -run '^TestDocuments(Migrations|KnowledgeFile|Recovery)AgainstPostgreSQL184$' -count=1
PASS (PostgreSQL 18.4; fresh and v1_adopted, 36.772s)
```

The full repo archguard exits non-zero with 28 lines containing a Documents
path or import. Twenty-one are frozen registration/boundary debt: six Knowledge
operation registrations, eight Knowledge SQL storage-directory claims, two
paired column-level permission parser findings and five Forms-record-archive
API/application boundary findings. Seven are Forms V1 compatibility imports of
the Documents archive API. None names the new production source or test files;
these remain isolated to the locked global integration pass.

## Single-owner integration work still pending

- Add migrations `20260801172000` and `20260802232400` to the global order once.
- Construct `NewConfiguredAPIProduction` once and inject its application,
  Knowledge-file and MCP direct owner interfaces. Schedule `ReconcileExpired`
  from the returned Knowledge-file owner.
- Construct `NewFormRecordArchiveOwner` once with the existing Documents archive
  IAM refresh and async job ports. Forms must call it only after its own
  record/field authorization and return the canonical `202` handle/location.
  HTTP, worker jobs, Forms links/cleanup and recovery are already composed and
  must not be reopened.
- Extend the typed worker profile with the canonical Documents ClamAV address,
  then construct `NewConfiguredWorkerProduction` once. Register its six job
  processors, Knowledge owner/expiry reconciler and MCP execution owner; do not
  also construct standalone Documents job resources.
- Register the thirteen direct and ten AI identities from the returned owner
  interfaces once.
- Register Knowledge operation/storage claims and normalize the column-level
  permission representation without widening database privileges.
- After every module is frozen, normalize permission/MCP/OpenAPI/parity
  manifests, migration order and global gate snapshots exactly once; classify
  the Forms linked ZIP route as `redesigned_equivalent` with this owner evidence.
- Finally prove the live Gateway path: login, tenant selection, IAM deny/allow,
  upload/download/archive, Forms cleanup, Knowledge lifecycle, purge,
  backup/restore replay and all Documents MCP operations.

No shared manifest, global migration registry, composition, OpenAPI, permission,
legacy snapshot or global gate file was edited by this module freeze.

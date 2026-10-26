# Forms record archive owner impact

Status: controlled Documents reopen complete; aggregate refrozen.

## Ownership

- Documents owns archive request state, pinned versions, job identity, encrypted
  artifact, expiry and purge receipts.
- Forms owns form-record read/field authorization and supplies the record's
  document IDs only after that decision.
- Documents verifies that the supplied IDs are the exact active
  `FORMS/form_record/<definition>/<record>/LIFECYCLE` link set before enqueue.
- Consumers use a new Documents `api/formrecordarchive` owner; no Documents
  application, adapter, table or object-store package crosses the boundary.

## Legacy evidence and compatibility

- V1 Gateway owns
  `GET /api/form-engine/dynamic/:slug/:id/documents/download`.
- The V1 Forms handler loads the record, extracts unique document IDs, queries
  document rows directly, reads MinIO objects and builds a synchronous in-memory
  ZIP. Missing objects are skipped and a response is returned when one survives.
- ADR-0017/0027 forbid that direct SQL/object access and synchronous V2 ZIP.
  The redesigned result is a `202` Documents archive handle followed by the
  canonical status/content routes. Entry names use Documents' deterministic
  archive naming rather than V1 field-label directories.

## Tenant, permission and data

- Tenant and actor come from opaque access evidence and are checked again by the
  Documents archive access authority before any tenant scope opens.
- Creation uses the existing billing+forms and documents authority generations,
  then Documents-owned link and archive adapters in one tenant command.
- Archive polling is actor-owned and refreshes `documents.item.download` before
  the Documents snapshot query.
- No schema, migration, backfill or second metadata authority is introduced.
  Existing `documents_archives`, `documents_archive_entries`,
  `documents_operations`, audit and `platform.jobs` rows remain authoritative.

## Async, idempotency and artifact lifecycle

- The existing archive command writes request, audit, operation receipt and job
  in the same tenant transaction.
- Caller idempotency is deterministically namespaced by definition and record;
  same source/key replays one handle and cross-record key reuse cannot alias it.
- Existing workers pin one ready version per item on first successful attempt,
  refresh authorization, stream an encrypted private artifact, retain it for
  24 hours and durably expire/purge it.

## Required verification

- Invalid actor/tenant/revision, denied IAM refresh and link substitution fail
  before archive creation.
- Exact links enqueue once; replay keeps the same derived key and handle.
- Cross-record key derivation differs deterministically.
- Status query is actor-bound and authorization-refreshed.
- Existing archive worker/artifact lifecycle suites remain green.
- Documents fileguard and targeted archguard remain green.

## Integration residue

- Forms must call the owner only after its record/field authorization and then
  return the canonical Documents archive handle/location.
- Composition must construct and inject the owner once after all affected
  modules are frozen. No global manifest/OpenAPI/parity edit occurs here.

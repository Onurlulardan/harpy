# Documents knowledge-file owner impact

Status: controlled reopen complete; see `freeze-handoff.md`.

## Ownership

- Documents remains the sole owner of bytes, versions, encryption, clean-scan
  evidence, upload state, lifecycle links and orphan cleanup.
- AI Knowledge owns semantic metadata, extracted chunks and index state only.
- Consumers depend on `api/knowledgefile`; no consumer imports Documents
  application, domain, PostgreSQL adapter, migration or table packages.

## Runtime boundary

- `Stage` reuses the existing folder and upload applications, including
  billing, scanner, encryption, object storage and upload reconciliation. It
  requires exact bytes/SHA-256 and retains a one-hour unadopted lease.
- `Adopt` and `Release` run in the caller's shared tenant transaction through
  Documents-owned link storage. The shared command requires real `ai` and
  `documents` owners; `platform_async` is not a participant.
- `Open` verifies the active AI lifecycle link and exact adopted item/version,
  then uses the standard Documents download pipeline.
- `ReconcileExpired` locks expired pending/staged leases and trashes each
  orphan through the standard Documents item application in the same tenant
  transaction. A delete failure rolls back the lease expiry mutation.
- Actor, tenant and authorization-version evidence are preserved across Stage
  and Open. Release permits a different currently authorized Knowledge manager
  while still pinning item, reference and source identities.

## Data and rollout

- Scope: tenant schema.
- New owner table: `documents_knowledge_file_leases`.
- `documents_links.source_owner` gains exact owner `AI`.
- Fresh and `v1_adopted` migration `20260801172000` are byte-identical.
- The additive table and released link evidence remain forward-repair data if
  a compatible V2 image is rolled back.

## Acceptance evidence

- Stage fails closed on byte, hash, scan, actor, tenant or MIME mismatch.
- Adopt/release pin lease, item, version, reference and source; replays are
  exact and substitution attempts fail.
- Open accepts only an active adopted exact version.
- Expired staged leases cannot be adopted and are selected with the current
  item row version for atomic standard trash.
- Fresh and adopted PostgreSQL 18.4 schemas execute the same owner lifecycle.

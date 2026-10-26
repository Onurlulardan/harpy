# System-resource layout strict-freeze handoff

Status: **MODULE-LOCAL FROZEN on 2026-08-02**. Reopen only for a reproducible
failure whose owner is this module. IAM and Catalog are frozen and were not
edited. Global manifests, aggregate OpenAPI, permission counts, migration
registries/frontiers, parity snapshots and global guard snapshots were not
edited.

This verdict is not local-stack or live-Gateway acceptance. Those remain the
final single-owner integration gate.

## Frozen owner boundary

- `system_resource_layout` owns presentation layout revisions, publications,
  reference evidence and installation-baseline evidence in each tenant schema.
- IAM remains the sole owner of user, role and organization values, field
  descriptors, mutability and authorization. The module consumes only the
  frozen IAM public field-catalog contract.
- Forms contributes only public presentation document/registry types. The
  module does not create Forms definitions, record tables, hooks or permissions.
- Every mutation runs inside the caller's verified `TenantScope` and binds only
  the layout PostgreSQL adapter.

## HTTP and authorization surface

The production composition already builds one handler with seven exact routes:

| Method | Path | IAM permission | Layout authority |
| --- | --- | --- | --- |
| GET | `/api/system-resources/{resource}/layout-revisions` | `iam.{resource}.read` | `system_resource_layout.list` |
| PUT | `/api/system-resources/{resource}/layout-revisions/{revision_id}` | `iam.{resource}.update` | `system_resource_layout.create` |
| GET | same revision path | `iam.{resource}.read` | `system_resource_layout.get` |
| PATCH | same revision path | `iam.{resource}.update` | `system_resource_layout.update` |
| DELETE | same revision path | `iam.{resource}.update` | `system_resource_layout.delete` |
| PUT | revision path plus `/publication` | `iam.{resource}.update` | `system_resource_layout.publish` |
| GET | `/api/system-resources/{resource}/presentation` | `iam.{resource}.read` | `system_resource_layout.resolve` |

Frozen IAM verifies bearer identity, operator access, exact tenant headers,
resource/action/operation pairs and equal positive `iam` plus
`system_resource_layout` generations before minting admission. Portal tokens,
cross-pairs, unknown operations and generation mismatch fail closed.

## IAM organization-owner parity closure

- Frozen IAM exposes `owner_ids` as field
  `c1000000-0000-4000-8000-00000000000c`, editable and required on create.
- Go default presentation includes it on organization create and update and
  keeps it visible on detail/list without turning layout into authorization.
- Installation baseline catalogs intentionally predate this field. The paired
  successor adds the exact `uuid_list_v1` editable catalog entry, creates a new
  immutable v2 definition/publication and advances generation 2 to 3.
- Custom user/role/organization publications are preserved. A corrupt
  deterministic curated predecessor now aborts with SQLSTATE `23514` rather
  than being silently treated as custom.
- Turkish canonical labels are stored as valid UTF-8 (`Oluştur`, `Güncelle`,
  `Ayrıntılar`); a regression test rejects their mojibake forms.

## Migration proof

- Selected logical versions: 47.
- Embedded documents: 94, exactly 47 fresh plus 47 `v1_adopted`.
- Selected stream span: `20260716013500..20260802232102`.
- New forward range: `20260802232100..20260802232102`.
- Fresh/adopted bytes are identical for every selected migration.
- Repository-wide read-only scan: zero version collision with another module.
- PostgreSQL repair tests no longer set
  `plpgsql.variable_conflict=use_variable`; the real fail-closed server default
  executes both origins successfully.

New paired SHA-256 values:

| Version | SHA-256 |
| --- | --- |
| `20260802232100` | `a1763e4380e2314851571cdad33226b6f4fe549674424e757f43ddf3d363b85e` |
| `20260802232101` | `d9ce76933eb2b36d87a605edf8fea05ddc7a4d07f095cfc36747cd0a382b117d` |
| `20260802232102` | `3ea486f86770ea9885271961a03c3fa57777018d9160f37d5adb14865c86ead8` |

## Verification evidence

Commands run from the `backend-v2` root:

| Check | Result |
| --- | --- |
| `go test ./core/internal/modules/system_resource_layout/... -count=1` | PASS |
| `go vet ./core/internal/modules/system_resource_layout/...` | PASS |
| `go run ./tools/fileguard -root core/internal/modules/system_resource_layout` | PASS |
| targeted presentation/layout archguard tests | PASS |
| full archguard read-only scan filtered to this module | zero findings; global gate still red elsewhere |
| system-layout orchestration unit suite | PASS |
| frozen IAM layout-access and owner-catalog target tests | PASS |
| `go build ./core/internal/composition/api` | PASS at the module snapshot; later concurrent Tenancy work made the shared package temporarily incomplete |
| canonical OpenAPI aggregate system-layout target | PASS |
| PG18.4 fresh/adopted full CRUD, publication, authority, isolation, rollback and concurrency | PASS |
| PG18.4 fresh/adopted installation/parity repair, replay, custom preservation and corruption denial | PASS |

The composition API test package changed concurrently during this freeze. One
run was blocked by a Forms fixture missing the newly added `DeleteV1Many` method;
the final rerun was blocked by in-progress Tenancy lifecycle routes whose
`list/get/update/suspend/reactivate/offboard/purge/readOperation` handlers were
not present yet. Production composition built before those Tenancy files landed,
and module/orchestration/IAM target tests pass. The final integration owner must
rerun the HTTP tests after Forms and Tenancy freeze; neither build error points
to System Layout source.

## Compatibility disposition for final parity inventory

- Legacy system form layouts for `users`, `roles` and `organizations` are
  `redesigned_equivalent`: the existing V1 definition edge projects fixed IDs
  and `is_system=true`, then delegates values/catalog to IAM and presentation to
  this module.
- Existing v1 layout bytes remain readable and immutable. Editing creates a v2
  successor; incompatible persisted evidence fails closed instead of falling
  back to a default.
- IAM resource delete remains forbidden. Layout revision deletion never deletes
  a user, role or organization.
- Route counts, compatibility routes and MCP capabilities must remain separate
  metrics in the one final global parity refresh.

## Exact single-owner integration residue

1. Register `20260802232100..20260802232102` once in the global migration
   manifest, release registry and normalized generation order. Do not renumber
   or rewrite the paired files.
2. Consume the new owner-aware catalog/publication digests once in final global
   migration and parity evidence. Do not reopen IAM or duplicate `owner_ids`.
3. Rerun the ten System Layout composition HTTP contract tests after the Forms
   fixture and in-progress Tenancy lifecycle build breaks are closed.
4. Normalize aggregate OpenAPI, operation/storage/permission and parity surfaces
   only in the final single-owner pass. Current full archguard has zero
   System Layout path finding but remains globally red.
5. Apply a clean fresh and adopted migration set, start the real stack and use
   Gateway port 8000 to live-test login, tenant selection, layout read/update/
   publish/resolve, IAM denial and V1 fixed system-form projection.

No module-local implementation, migration, permission or targeted-test item
remains open.

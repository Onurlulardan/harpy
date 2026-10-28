# Migration bundle impact

Owner: platform migration bundle (`internal/migrationbundle`, fresh manifest producers, startup contract binding).

Affected contracts: fresh migration manifest bytes and the existing binary schema contract digest. Release attestation JSON files are not rewritten. Goose stays.

Affected data/migrations: no schema SQL, no applied history rewrite, no tenant table change. Fresh authoring still stamps `expand` / database generation 1 / runtime generation 1. Archived manifests keep their stored phase and generation.

Tenant and permission risk: none. Tenant id still comes from the actor context. This work does not authorize requests or read another tenant's rows.

Events/jobs: none.

V1 compatibility: fresh manifests still carry only the fresh variant. Adopted paths remain inside archived release manifests. The 21 fresh/adopted SHA mismatches stay as recorded history. Adopted SQL is not in this tree.

Rollout and rollback: API and worker startup read the embedded fresh manifest and require its SHA-256 to match `binary-schema-contract.json`. Migrator and operator load `release-index.json` and read one historical manifest from `migration-release.zip` or the source release directory. Roll back by restoring the full manifest embed in `LoadReleaseRegistry` and removing the zip copy from the production image.

Required tests: producer byte equality, determinism, digest mismatch, history supersession chain preservation, filesystem registry equals embed, bound startup reject paths.

Open decisions: fresh authoring still stamps expand / generation 1 / 1. Adopted SQL dedup stays out because 21 entries differ and rewriting manifests would change attestation bytes.

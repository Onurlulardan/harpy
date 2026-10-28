# Migration bundle impact

Owner: platform migration bundle (`internal/migrationbundle`, fresh manifest producers, startup contract binding).

Affected contracts: fresh migration manifest bytes and the existing binary schema contract digest. Release attestation JSON files are not rewritten. Goose stays.

Affected data/migrations: no schema SQL, no applied history rewrite, no tenant table change. Fresh authoring still stamps `expand` / database generation 1 / runtime generation 1. Archived manifests keep their stored phase and generation.

Tenant and permission risk: none. Tenant id still comes from the actor context. This work does not authorize requests or read another tenant's rows.

Events/jobs: none.

V1 compatibility: fresh manifests still carry only the fresh variant. Adopted paths remain inside archived release manifests. The 21 fresh/adopted SHA mismatches stay as recorded history. Adopted SQL is not in this tree.

Rollout and rollback: API and worker startup read the embedded fresh manifest and require its SHA-256 to match `binary-schema-contract.json`. The SQL corpus check remains in tests. Historical release manifests stay embedded for migrator and operator authorization. Roll back by restoring the previous startup loader and deleting the fresh manifest artifact.

Required tests: producer byte equality, determinism, digest mismatch, history supersession chain preservation, filesystem registry equals embed, bound startup reject paths.

Open decisions: do not remove the historical embed. The distroless image ships binaries, not the release directory, and `Authorize` still requires canonical bytes.

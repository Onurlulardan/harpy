INSERT INTO platform.tenant_migration_requirement_catalogs (
    migration_run_id, contract_version, release_id, environment_id,
    database_origin, migration_manifest_sha256, catalog_sha256,
    requirement_count, target_count, canonical_document
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (migration_run_id) DO NOTHING;

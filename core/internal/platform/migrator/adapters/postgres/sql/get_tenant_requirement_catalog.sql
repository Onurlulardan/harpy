SELECT contract_version, migration_run_id::text, release_id::text,
       environment_id::text, database_origin, migration_manifest_sha256,
       catalog_sha256, requirement_count, target_count, canonical_document, recorded_at
FROM platform.tenant_migration_requirement_catalogs
WHERE migration_run_id = $1;

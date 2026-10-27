SELECT ordinal, owner, logical_version, database_origin,
       selected_path, selected_sha256
FROM platform.tenant_migration_requirements
WHERE migration_run_id = $1
ORDER BY ordinal;

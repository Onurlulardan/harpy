SELECT ordinal, tenant_id::text, schema_name
FROM platform.tenant_migration_requirement_targets
WHERE migration_run_id = $1
ORDER BY ordinal;

INSERT INTO platform.tenant_migration_requirement_targets (
    migration_run_id, ordinal, tenant_id, schema_name
)
SELECT $1, target.ordinal, target.tenant_id, target.schema_name
FROM pg_catalog.jsonb_to_recordset($2::jsonb) AS target(
    ordinal integer, tenant_id uuid, schema_name text
);

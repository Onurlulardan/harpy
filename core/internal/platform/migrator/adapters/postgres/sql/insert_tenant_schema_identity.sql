INSERT INTO tenancy_schema_identity (
    singleton, tenant_id, schema_name, origin, created_at
) VALUES (true, $1, $2, $3, pg_catalog.clock_timestamp())
ON CONFLICT (singleton) DO NOTHING;

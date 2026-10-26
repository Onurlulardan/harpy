SELECT tenant_id, schema_name, namespace_exists, schema_receipt_exists
FROM platform.tenancy_tenant_migration_schema_inspect_v1($1, $2, $3, $4)

SELECT tenant_id::text, schema_name, origin
FROM tenancy_schema_identity
WHERE singleton = true;

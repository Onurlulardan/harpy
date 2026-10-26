SELECT tenant_id::text, schema_name
FROM tenancy_schema_identity
WHERE singleton = TRUE
  AND tenant_id = $1
  AND schema_name = $2;

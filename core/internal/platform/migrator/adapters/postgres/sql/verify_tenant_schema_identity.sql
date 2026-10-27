SELECT count(*) = 1
FROM tenancy_schema_identity
WHERE singleton = true AND tenant_id = $1 AND schema_name = $2 AND origin = $3;

SELECT catalog_sha256, requirement_count
FROM platform.tenancy_tenant_migration_complete_v3($1, $2, $3, $4, $5, $6, $7)

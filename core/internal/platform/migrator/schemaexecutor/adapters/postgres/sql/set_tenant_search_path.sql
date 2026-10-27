SELECT
    pg_catalog.set_config('search_path', $1, true),
    tenant.id::text
FROM platform.tenants AS tenant
WHERE tenant.id = $2;

SELECT tenant.id::text, tenant.schema_name
FROM platform.tenants AS tenant
WHERE tenant.deleted_at IS NULL
ORDER BY tenant.id;

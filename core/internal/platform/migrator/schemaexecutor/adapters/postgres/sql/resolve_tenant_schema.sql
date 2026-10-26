SELECT
    tenant.schema_name,
    tenant.status,
    tenant.deleted_at IS NULL,
    (
        SELECT namespace.nspowner::regrole::text
        FROM pg_catalog.pg_namespace AS namespace
        WHERE namespace.nspname = tenant.schema_name
    )
FROM platform.tenants AS tenant
WHERE tenant.id = $1
FOR SHARE;

SELECT
    tenant.id::text,
    pg_catalog.pg_advisory_unlock(
        pg_catalog.hashtextextended(
            'harpy:v2:migrate:tenant:' || tenant.id::text,
            0
        )
    )
FROM platform.tenants AS tenant
WHERE tenant.id = $1;

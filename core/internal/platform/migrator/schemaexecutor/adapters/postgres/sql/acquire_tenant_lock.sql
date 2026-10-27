WITH locked AS MATERIALIZED (
    SELECT
        tenant.id,
        pg_catalog.pg_advisory_lock(
        pg_catalog.hashtextextended(
            'harpy:v2:migrate:tenant:' || tenant.id::text,
            0
        )
        ) AS acquired
    FROM platform.tenants AS tenant
    WHERE tenant.id = $1
)
SELECT locked.id::text
FROM locked;

SELECT target.environment_id::text,
       target.deployment_mode,
       target.tenant_id::text,
       target.purpose,
       target.slug,
       target.environment_id IS NOT NULL
FROM (VALUES (true)) AS singleton(seed)
LEFT JOIN LATERAL (
    SELECT * FROM platform.tenancy_get_primary_tenant_target_v1()
) AS target ON singleton.seed;

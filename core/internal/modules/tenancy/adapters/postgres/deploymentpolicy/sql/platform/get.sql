SELECT policy.environment_id::text,
       policy.deployment_mode,
       policy.maximum_customer_tenants,
       policy.created_at,
       policy.environment_id IS NOT NULL
FROM (VALUES (true)) AS singleton(seed)
LEFT JOIN LATERAL (
    SELECT * FROM platform.tenancy_get_deployment_policy_v1()
) AS policy ON singleton.seed;

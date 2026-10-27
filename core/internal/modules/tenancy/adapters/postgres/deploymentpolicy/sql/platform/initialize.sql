SELECT operation_id, environment_id, deployment_mode,
       maximum_customer_tenants, created_at, database_now, replayed
FROM platform.tenancy_initialize_deployment_policy_v1(
    $1::uuid, $2::uuid, $3::text, $4::text
);

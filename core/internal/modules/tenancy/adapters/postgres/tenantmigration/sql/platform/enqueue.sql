SELECT job_id::text, operation_id::text, state, tenant_id::text,
       schema_name, step_plan_sha256, $3::uuid::text AS release_id,
       $4::bigint AS runtime_generation, $5::bigint AS database_generation,
       $6::text AS migration_manifest_sha256, NULL::text AS error_code
FROM platform.tenancy_tenant_migration_enqueue_v1($1, $2, $3, $4, $5, $6)

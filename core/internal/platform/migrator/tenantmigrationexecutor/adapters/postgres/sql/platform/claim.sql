SELECT job_id, operation_id, tenant_id, schema_name, step_plan_sha256,
       origin, attempt, migration_run_id, previous_run_id, lease_id,
       profile_instance_id, release_id, environment_id, deployment_mode,
       runtime_generation, database_generation, migration_manifest_sha256,
       lease_until
FROM platform.tenancy_tenant_migration_claim_v1($1, $2, $3, $4, $5)

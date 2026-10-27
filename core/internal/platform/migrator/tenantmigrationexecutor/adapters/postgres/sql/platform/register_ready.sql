SELECT profile_instance_id, release_id, environment_id, deployment_mode,
       runtime_generation, database_generation, migration_manifest_sha256,
       ready_until
FROM platform.tenancy_tenant_migration_ready_register_v1($1, $2, $3, $4, $5, $6)

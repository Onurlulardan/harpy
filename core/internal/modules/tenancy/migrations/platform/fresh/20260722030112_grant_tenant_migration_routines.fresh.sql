-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_ready_current_v1()
RETURNS TABLE (
    release_id uuid,
    environment_id uuid,
    deployment_mode text,
    runtime_generation bigint,
    database_generation bigint,
    migration_manifest_sha256 text,
    ready_until timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN QUERY
    SELECT ready.release_id, ready.environment_id, ready.deployment_mode,
           ready.runtime_generation, ready.database_generation,
           ready.migration_manifest_sha256, ready.ready_until
    FROM platform.tenancy_tenant_migration_executor_readiness AS ready
    JOIN platform.tenancy_deployment_policy AS policy
      ON policy.environment_id = ready.environment_id
     AND policy.deployment_mode = ready.deployment_mode
    JOIN platform.release_generations AS release_row
      ON release_row.environment_id = ready.environment_id
     AND release_row.release_id = ready.release_id
     AND release_row.runtime_generation = ready.runtime_generation
     AND release_row.database_generation = ready.database_generation
     AND release_row.migration_manifest_sha256 =
         ready.migration_manifest_sha256
    JOIN platform.schema_contracts AS contract
      ON contract.component = 'core' AND contract.status = 'ready'
     AND contract.verified_run_id = release_row.migration_run_id
    WHERE ready.state = 'active'
      AND ready.ready_until > pg_catalog.clock_timestamp()
    ORDER BY ready.ready_until DESC, ready.profile_instance_id
    LIMIT 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_ready_current_v1()
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_tenant_migration_ready_current_v1()
TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_tenant_migration_enqueue_v1(
    uuid, uuid, uuid, bigint, bigint, text
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.tenancy_tenant_migration_get_v1(uuid)
TO harpy_runtime;

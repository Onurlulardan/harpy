-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_ready_register_v1(
    p_profile_instance_id uuid,
    p_release_id uuid,
    p_deployment_mode text,
    p_runtime_generation bigint,
    p_migration_manifest_sha256 text,
    p_ttl_seconds integer
)
RETURNS TABLE (
    profile_instance_id uuid,
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
DECLARE
    v_environment_id uuid;
    v_database_generation bigint;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_profile_instance_id IS NULL OR p_release_id IS NULL OR
       p_deployment_mode NOT IN ('onpremise', 'saas') OR
       p_runtime_generation < 1 OR
       p_migration_manifest_sha256 !~ '^[0-9a-f]{64}$' OR
       p_ttl_seconds NOT BETWEEN 15 AND 120 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenant migration readiness request is invalid';
    END IF;

    SELECT policy.environment_id, release_row.database_generation
    INTO v_environment_id, v_database_generation
    FROM platform.tenancy_deployment_policy AS policy
    JOIN platform.release_generations AS release_row
      ON release_row.environment_id = policy.environment_id
    JOIN platform.schema_contracts AS contract
      ON contract.component = 'core'
     AND contract.verified_run_id = release_row.migration_run_id
     AND contract.status = 'ready'
     AND contract.database_generation = release_row.database_generation
     AND contract.migration_manifest_sha256 =
         release_row.migration_manifest_sha256
    WHERE policy.deployment_mode = p_deployment_mode
      AND release_row.release_id = p_release_id
      AND release_row.runtime_generation = p_runtime_generation
      AND release_row.migration_manifest_sha256 =
          p_migration_manifest_sha256
    FOR SHARE OF policy, release_row, contract;

    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration release is not current';
    END IF;

    INSERT INTO platform.tenancy_tenant_migration_executor_readiness (
        profile_instance_id, release_id, environment_id, deployment_mode,
        runtime_generation, database_generation, migration_manifest_sha256,
        state, ready_until, heartbeat_at, created_at
    ) VALUES (
        p_profile_instance_id, p_release_id, v_environment_id,
        p_deployment_mode, p_runtime_generation, v_database_generation,
        p_migration_manifest_sha256, 'active',
        v_now + pg_catalog.make_interval(secs => p_ttl_seconds), v_now, v_now
    )
    ON CONFLICT ON CONSTRAINT
        tenancy_tenant_migration_executor_readiness_pkey DO UPDATE SET
        heartbeat_at = EXCLUDED.heartbeat_at,
        ready_until = EXCLUDED.ready_until
    WHERE tenancy_tenant_migration_executor_readiness.state = 'active'
      AND tenancy_tenant_migration_executor_readiness.release_id =
          EXCLUDED.release_id
      AND tenancy_tenant_migration_executor_readiness.environment_id =
          EXCLUDED.environment_id
      AND tenancy_tenant_migration_executor_readiness.deployment_mode =
          EXCLUDED.deployment_mode
      AND tenancy_tenant_migration_executor_readiness.runtime_generation =
          EXCLUDED.runtime_generation
      AND tenancy_tenant_migration_executor_readiness.database_generation =
          EXCLUDED.database_generation
      AND tenancy_tenant_migration_executor_readiness.migration_manifest_sha256 =
          EXCLUDED.migration_manifest_sha256;

    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration profile instance cannot be reused';
    END IF;

    RETURN QUERY
    SELECT ready.profile_instance_id, ready.release_id, ready.environment_id,
           ready.deployment_mode, ready.runtime_generation,
           ready.database_generation, ready.migration_manifest_sha256,
           ready.ready_until
    FROM platform.tenancy_tenant_migration_executor_readiness AS ready
    WHERE ready.profile_instance_id = p_profile_instance_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_ready_register_v1(
    uuid, uuid, text, bigint, text, integer
) FROM PUBLIC;

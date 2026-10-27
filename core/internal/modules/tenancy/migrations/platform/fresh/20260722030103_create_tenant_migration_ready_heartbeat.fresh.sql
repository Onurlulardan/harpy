-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_ready_heartbeat_v1(
    p_profile_instance_id uuid,
    p_ttl_seconds integer
)
RETURNS timestamptz
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_ready_until timestamptz;
BEGIN
    IF p_profile_instance_id IS NULL OR p_ttl_seconds NOT BETWEEN 15 AND 120 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenant migration heartbeat request is invalid';
    END IF;

    UPDATE platform.tenancy_tenant_migration_executor_readiness AS ready
    SET heartbeat_at = v_now,
        ready_until = v_now + pg_catalog.make_interval(secs => p_ttl_seconds)
    FROM platform.tenancy_deployment_policy AS policy,
         platform.release_generations AS release_row,
         platform.schema_contracts AS contract
    WHERE ready.profile_instance_id = p_profile_instance_id
      AND ready.state = 'active'
      AND ready.ready_until > v_now
      AND policy.environment_id = ready.environment_id
      AND policy.deployment_mode = ready.deployment_mode
      AND release_row.environment_id = ready.environment_id
      AND release_row.release_id = ready.release_id
      AND release_row.runtime_generation = ready.runtime_generation
      AND release_row.database_generation = ready.database_generation
      AND release_row.migration_manifest_sha256 =
          ready.migration_manifest_sha256
      AND contract.component = 'core'
      AND contract.status = 'ready'
      AND contract.verified_run_id = release_row.migration_run_id
      AND contract.database_generation = ready.database_generation
      AND contract.migration_manifest_sha256 =
          ready.migration_manifest_sha256
    RETURNING ready.ready_until INTO v_ready_until;

    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration readiness is stale';
    END IF;
    RETURN v_ready_until;
END;
$function$;

CREATE FUNCTION platform.tenancy_tenant_migration_ready_retire_v1(
    p_profile_instance_id uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    UPDATE platform.tenancy_tenant_migration_executor_readiness
    SET state = 'retired', retired_at = v_now
    WHERE profile_instance_id = p_profile_instance_id AND state = 'active';
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_ready_heartbeat_v1(
    uuid, integer
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_ready_retire_v1(uuid)
FROM PUBLIC;

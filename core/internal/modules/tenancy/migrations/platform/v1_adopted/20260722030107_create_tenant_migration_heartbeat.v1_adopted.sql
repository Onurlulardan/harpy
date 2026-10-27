-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_heartbeat_v1(
    p_job_id uuid,
    p_attempt integer,
    p_profile_instance_id uuid,
    p_lease_id uuid,
    p_lease_seconds integer
)
RETURNS timestamptz
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_lease_until timestamptz;
BEGIN
    IF p_job_id IS NULL OR p_attempt NOT BETWEEN 1 AND 8 OR
       p_profile_instance_id IS NULL OR p_lease_id IS NULL OR
       p_lease_seconds NOT BETWEEN 30 AND 300 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenant migration lease heartbeat is invalid';
    END IF;

    UPDATE platform.tenancy_tenant_migration_attempts AS attempt_row
    SET heartbeat_at = v_now,
        lease_until = v_now + pg_catalog.make_interval(secs => p_lease_seconds)
    FROM platform.tenancy_tenant_migration_jobs AS job,
         platform.tenancy_tenant_migration_executor_readiness AS ready
    WHERE attempt_row.job_id = p_job_id
      AND attempt_row.attempt = p_attempt
      AND attempt_row.profile_instance_id = p_profile_instance_id
      AND attempt_row.lease_id = p_lease_id
      AND attempt_row.state = 'running'
      AND attempt_row.lease_until > v_now
      AND job.job_id = attempt_row.job_id
      AND job.state = 'running'
      AND job.current_attempt = attempt_row.attempt
      AND ready.profile_instance_id = p_profile_instance_id
      AND ready.state = 'active'
      AND ready.ready_until > v_now
      AND ready.release_id = job.release_id
      AND ready.environment_id = job.environment_id
      AND ready.runtime_generation = job.runtime_generation
      AND ready.database_generation = job.database_generation
      AND ready.migration_manifest_sha256 = job.migration_manifest_sha256
    RETURNING attempt_row.lease_until INTO v_lease_until;

    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration lease is stale';
    END IF;
    RETURN v_lease_until;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_heartbeat_v1(
    uuid, integer, uuid, uuid, integer
) FROM PUBLIC;

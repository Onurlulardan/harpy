-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_enqueue_v1(
    p_job_id uuid,
    p_operation_id uuid,
    p_release_id uuid,
    p_runtime_generation bigint,
    p_database_generation bigint,
    p_migration_manifest_sha256 text
)
RETURNS TABLE (
    job_id uuid,
    operation_id uuid,
    state text,
    tenant_id uuid,
    schema_name text,
    step_plan_sha256 text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.tenancy_onboarding_operations%ROWTYPE;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_job_id IS NULL OR p_operation_id IS NULL OR p_release_id IS NULL OR
       p_runtime_generation < 1 OR p_database_generation < 1 OR
       p_migration_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenant migration enqueue request is invalid';
    END IF;

    SELECT * INTO v_operation
    FROM platform.tenancy_onboarding_operations
    WHERE tenancy_onboarding_operations.operation_id = p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR v_operation.state NOT IN ('accepted', 'migrating') THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'onboarding operation cannot enqueue migration';
    END IF;

    IF NOT EXISTS (
        SELECT 1
        FROM platform.tenancy_tenant_migration_executor_readiness AS ready
        WHERE ready.release_id = p_release_id
          AND ready.environment_id = v_operation.environment_id
          AND ready.deployment_mode = v_operation.deployment_mode
          AND ready.runtime_generation = p_runtime_generation
          AND ready.database_generation = p_database_generation
          AND ready.migration_manifest_sha256 =
              p_migration_manifest_sha256
          AND ready.state = 'active'
          AND ready.ready_until > v_now
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration executor is not claim ready';
    END IF;

    INSERT INTO platform.tenancy_tenant_migration_jobs (
        job_id, operation_id, release_id, environment_id, deployment_mode,
        runtime_generation, database_generation, migration_manifest_sha256,
        state, created_at, updated_at
    ) VALUES (
        p_job_id, p_operation_id, p_release_id, v_operation.environment_id,
        v_operation.deployment_mode, p_runtime_generation,
        p_database_generation, p_migration_manifest_sha256,
        'pending', v_now, v_now
    ) ON CONFLICT ON CONSTRAINT
        tenancy_tenant_migration_jobs_operation_id_key DO NOTHING;

    IF v_operation.state = 'accepted' THEN
        UPDATE platform.tenancy_onboarding_operations AS operation_row
        SET state = 'migrating', revision = operation_row.revision + 1,
            updated_at = v_now
        WHERE operation_row.operation_id = p_operation_id
          AND operation_row.state = 'accepted';
    END IF;

    RETURN QUERY
    SELECT job.job_id, job.operation_id, job.state,
           operation.tenant_id, operation.schema_name::text,
           operation.step_plan_sha256
    FROM platform.tenancy_tenant_migration_jobs AS job
    JOIN platform.tenancy_onboarding_operations AS operation
      ON operation.operation_id = job.operation_id
    WHERE job.operation_id = p_operation_id
      AND job.release_id = p_release_id
      AND job.runtime_generation = p_runtime_generation
      AND job.database_generation = p_database_generation
      AND job.migration_manifest_sha256 = p_migration_manifest_sha256;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration enqueue replay differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_enqueue_v1(
    uuid, uuid, uuid, bigint, bigint, text
) FROM PUBLIC;

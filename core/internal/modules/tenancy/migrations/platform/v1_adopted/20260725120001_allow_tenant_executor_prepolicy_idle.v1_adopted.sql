-- +goose Up
DROP FUNCTION platform.tenancy_tenant_migration_claim_v1(
    uuid, uuid, uuid, uuid, integer
);
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_claim_v1(
    p_profile_instance_id uuid,
    p_lease_id uuid,
    p_migration_run_id uuid,
    p_expiry_receipt_id uuid,
    p_lease_seconds integer
)
RETURNS TABLE (
    job_id uuid, operation_id uuid, tenant_id uuid, schema_name text,
    step_plan_sha256 text, origin text, attempt integer,
    migration_run_id uuid, previous_run_id uuid, lease_id uuid,
    profile_instance_id uuid, release_id uuid, environment_id uuid,
    deployment_mode text, runtime_generation bigint,
    database_generation bigint, migration_manifest_sha256 text,
    lease_until timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_ready platform.tenancy_tenant_migration_executor_readiness%ROWTYPE;
    v_job platform.tenancy_tenant_migration_jobs%ROWTYPE;
    v_previous_run_id uuid;
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_attempt integer;
    v_digest text;
BEGIN
    IF p_profile_instance_id IS NULL OR p_lease_id IS NULL OR
       p_migration_run_id IS NULL OR p_expiry_receipt_id IS NULL OR
       p_lease_seconds NOT BETWEEN 30 AND 300 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenant migration claim request is invalid';
    END IF;

    SELECT ready.* INTO v_ready
    FROM platform.tenancy_tenant_migration_executor_readiness AS ready
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
    WHERE ready.profile_instance_id = p_profile_instance_id
      AND ready.state = 'active' AND ready.ready_until > v_now
    FOR SHARE OF ready, release_row, contract;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration executor is not claim ready';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.tenancy_deployment_policy AS policy
        WHERE policy.environment_id IS DISTINCT FROM v_ready.environment_id
           OR policy.deployment_mode IS DISTINCT FROM v_ready.deployment_mode
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration deployment policy conflicts';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.tenancy_deployment_policy) THEN
        RETURN;
    END IF;

    UPDATE platform.tenancy_tenant_migration_jobs AS exhausted_job
    SET state = 'fenced', error_code = 'attempt_limit_exhausted',
        completed_at = v_now, updated_at = v_now
    WHERE exhausted_job.state = 'failed' AND exhausted_job.current_attempt >= 8
      AND exhausted_job.release_id = v_ready.release_id
      AND exhausted_job.environment_id = v_ready.environment_id
      AND exhausted_job.runtime_generation = v_ready.runtime_generation
      AND exhausted_job.database_generation = v_ready.database_generation
      AND exhausted_job.migration_manifest_sha256 = v_ready.migration_manifest_sha256;

    SELECT job.* INTO v_job
    FROM platform.tenancy_tenant_migration_jobs AS job
    WHERE (job.state IN ('pending', 'failed') OR (
        job.state = 'running' AND EXISTS (
            SELECT 1
            FROM platform.tenancy_tenant_migration_attempts AS attempt_row
            WHERE attempt_row.job_id = job.job_id
              AND attempt_row.attempt = job.current_attempt
              AND attempt_row.state = 'running'
              AND attempt_row.lease_until <= v_now
        )
    ))
      AND job.current_attempt < 8
      AND job.release_id = v_ready.release_id
      AND job.environment_id = v_ready.environment_id
      AND job.deployment_mode = v_ready.deployment_mode
      AND job.runtime_generation = v_ready.runtime_generation
      AND job.database_generation = v_ready.database_generation
      AND job.migration_manifest_sha256 = v_ready.migration_manifest_sha256
    ORDER BY job.created_at, job.job_id
    FOR UPDATE SKIP LOCKED
    LIMIT 1;
    IF NOT FOUND THEN
        RETURN;
    END IF;

    IF v_job.state = 'running' THEN
        SELECT attempt_row.migration_run_id INTO v_previous_run_id
        FROM platform.tenancy_tenant_migration_attempts AS attempt_row
        WHERE attempt_row.job_id = v_job.job_id
          AND attempt_row.attempt = v_job.current_attempt
        FOR UPDATE;
        UPDATE platform.tenancy_tenant_migration_attempts
        SET state = 'failed', finished_at = v_now,
            error_code = 'lease_expired'
        WHERE tenancy_tenant_migration_attempts.job_id = v_job.job_id
          AND tenancy_tenant_migration_attempts.attempt = v_job.current_attempt;
        v_digest := pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
            v_job.job_id::text || ':' || v_job.current_attempt::text ||
            ':lease_expired', 'UTF8')), 'hex');
        INSERT INTO platform.tenancy_tenant_migration_receipts (
            receipt_id, job_id, attempt, receipt_kind, migration_run_id,
            receipt_sha256, evidence_sha256, occurred_at
        ) VALUES (
            p_expiry_receipt_id, v_job.job_id, v_job.current_attempt,
            'failure', v_previous_run_id, v_digest, v_digest, v_now
        ) ON CONFLICT ON CONSTRAINT
            tenancy_tenant_migration_receipt_attempt_kind_uq
          DO NOTHING;
    END IF;

    v_attempt := v_job.current_attempt + 1;
    UPDATE platform.tenancy_tenant_migration_jobs
    SET state = 'running', current_attempt = v_attempt, error_code = NULL,
        completed_at = NULL, updated_at = v_now
    WHERE tenancy_tenant_migration_jobs.job_id = v_job.job_id;
    INSERT INTO platform.tenancy_tenant_migration_attempts (
        job_id, attempt, migration_run_id, profile_instance_id, lease_id,
        state, leased_at, heartbeat_at, lease_until
    ) VALUES (
        v_job.job_id, v_attempt, p_migration_run_id, p_profile_instance_id,
        p_lease_id, 'running', v_now, v_now,
        v_now + pg_catalog.make_interval(secs => p_lease_seconds)
    );

    RETURN QUERY
    SELECT v_job.job_id, v_job.operation_id, operation.tenant_id,
           operation.schema_name::text, operation.step_plan_sha256,
           v_job.origin, v_attempt, p_migration_run_id, v_previous_run_id,
           p_lease_id, p_profile_instance_id, v_job.release_id,
           v_job.environment_id, v_job.deployment_mode,
           v_job.runtime_generation, v_job.database_generation,
           v_job.migration_manifest_sha256,
           v_now + pg_catalog.make_interval(secs => p_lease_seconds)
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id = v_job.operation_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_claim_v1(
    uuid, uuid, uuid, uuid, integer
) FROM PUBLIC;

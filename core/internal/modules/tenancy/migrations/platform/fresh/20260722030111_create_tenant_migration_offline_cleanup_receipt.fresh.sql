-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_offline_cleanup_record_v1(
    p_job_id uuid,
    p_receipt_id uuid,
    p_receipt_sha256 text,
    p_evidence_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_job platform.tenancy_tenant_migration_jobs%ROWTYPE;
    v_schema_name text;
    v_tenant_id uuid;
    v_run_id uuid;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_job_id IS NULL OR p_receipt_id IS NULL OR
       p_receipt_sha256 !~ '^[0-9a-f]{64}$' OR
       p_evidence_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'offline cleanup receipt is invalid';
    END IF;
    SELECT job.* INTO v_job
    FROM platform.tenancy_tenant_migration_jobs AS job
    WHERE job.job_id = p_job_id
    FOR UPDATE;
    IF NOT FOUND OR v_job.state <> 'cleanup_required' OR EXISTS (
        SELECT 1
        FROM platform.tenancy_tenant_migration_executor_readiness AS ready
        WHERE ready.environment_id = v_job.environment_id
          AND ready.release_id = v_job.release_id
          AND ready.state = 'active' AND ready.ready_until > v_now
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'offline cleanup requires stopped executor';
    END IF;
    SELECT operation.schema_name::text, operation.tenant_id,
           attempt_row.migration_run_id
    INTO v_schema_name, v_tenant_id, v_run_id
    FROM platform.tenancy_onboarding_operations AS operation
    JOIN platform.tenancy_tenant_migration_attempts AS attempt_row
      ON attempt_row.job_id = v_job.job_id
     AND attempt_row.attempt = v_job.current_attempt
    WHERE operation.operation_id = v_job.operation_id;
    IF pg_catalog.to_regnamespace(v_schema_name) IS NOT NULL OR EXISTS (
        SELECT 1 FROM platform.tenants WHERE id = v_tenant_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'offline cleanup has not removed orphan schema';
    END IF;
    INSERT INTO platform.tenancy_tenant_migration_receipts (
        receipt_id, job_id, attempt, receipt_kind, migration_run_id,
        receipt_sha256, evidence_sha256, occurred_at
    ) VALUES (
        p_receipt_id, v_job.job_id, v_job.current_attempt,
        'offline_cleanup_completed', v_run_id,
        p_receipt_sha256, p_evidence_sha256, v_now
    );
    UPDATE platform.tenancy_tenant_migration_jobs
    SET state = 'fenced', error_code = 'offline_cleanup_completed',
        updated_at = v_now
    WHERE job_id = v_job.job_id;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.tenancy_tenant_migration_offline_cleanup_record_v1(
    uuid, uuid, text, text
) FROM PUBLIC;

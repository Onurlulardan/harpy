-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_fail_v1(
    p_job_id uuid,
    p_attempt integer,
    p_profile_instance_id uuid,
    p_lease_id uuid,
    p_outcome text,
    p_error_code text,
    p_receipt_id uuid,
    p_receipt_sha256 text,
    p_evidence_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_run_id uuid;
    v_receipt_kind text;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_outcome NOT IN ('failed', 'reconcile_required', 'cleanup_required') OR
       p_error_code !~ '^[a-z][a-z0-9_]{2,63}$' OR p_receipt_id IS NULL OR
       p_receipt_sha256 !~ '^[0-9a-f]{64}$' OR
       p_evidence_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenant migration failure receipt is invalid';
    END IF;
    SELECT attempt_row.migration_run_id INTO v_run_id
    FROM platform.tenancy_tenant_migration_jobs AS job
    JOIN platform.tenancy_tenant_migration_attempts AS attempt_row
      ON attempt_row.job_id = job.job_id
     AND attempt_row.attempt = job.current_attempt
    WHERE job.job_id = p_job_id AND job.state = 'running'
      AND job.current_attempt = p_attempt
      AND attempt_row.profile_instance_id = p_profile_instance_id
      AND attempt_row.lease_id = p_lease_id
      AND attempt_row.state = 'running'
    FOR UPDATE OF job, attempt_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration failure lease is stale';
    END IF;

    v_receipt_kind := CASE p_outcome
        WHEN 'failed' THEN 'failure'
        ELSE p_outcome
    END;
    INSERT INTO platform.tenancy_tenant_migration_receipts (
        receipt_id, job_id, attempt, receipt_kind, migration_run_id,
        receipt_sha256, evidence_sha256, occurred_at
    ) VALUES (
        p_receipt_id, p_job_id, p_attempt, v_receipt_kind, v_run_id,
        p_receipt_sha256, p_evidence_sha256, v_now
    );
    UPDATE platform.tenancy_tenant_migration_attempts
    SET state = p_outcome, finished_at = v_now, error_code = p_error_code
    WHERE job_id = p_job_id AND attempt = p_attempt;
    UPDATE platform.tenancy_tenant_migration_jobs
    SET state = p_outcome, error_code = p_error_code,
        completed_at = v_now, updated_at = v_now
    WHERE job_id = p_job_id;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_fail_v1(
    uuid, integer, uuid, uuid, text, text, uuid, text, text
) FROM PUBLIC;

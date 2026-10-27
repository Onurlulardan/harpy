-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_tenant_migration_complete_v1(
    p_job_id uuid,
    p_attempt integer,
    p_profile_instance_id uuid,
    p_lease_id uuid,
    p_receipt_id uuid,
    p_receipt_sha256 text,
    p_evidence_sha256 text
)
RETURNS TABLE (catalog_sha256 text, requirement_count integer)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_job platform.tenancy_tenant_migration_jobs%ROWTYPE;
    v_operation platform.tenancy_onboarding_operations%ROWTYPE;
    v_run_id uuid;
    v_origin text;
    v_verified boolean;
    v_catalog_sha256 text;
    v_requirement_count integer;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_receipt_id IS NULL OR p_receipt_sha256 !~ '^[0-9a-f]{64}$' OR
       p_evidence_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenant migration completion receipt is invalid';
    END IF;
    SELECT job.* INTO v_job
    FROM platform.tenancy_tenant_migration_jobs AS job
    WHERE job.job_id = p_job_id
    FOR UPDATE;
    IF NOT FOUND OR v_job.state <> 'running' OR
       v_job.current_attempt <> p_attempt THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration completion state is invalid';
    END IF;
    SELECT operation.* INTO v_operation
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id = v_job.operation_id
    FOR UPDATE;
    IF NOT FOUND OR v_operation.state <> 'migrating' OR EXISTS (
           SELECT 1 FROM platform.tenants WHERE id = v_operation.tenant_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration completion state is invalid';
    END IF;

    SELECT attempt_row.migration_run_id
    INTO v_run_id
    FROM platform.tenancy_tenant_migration_attempts AS attempt_row
    WHERE attempt_row.job_id = p_job_id AND attempt_row.attempt = p_attempt
      AND attempt_row.profile_instance_id = p_profile_instance_id
      AND attempt_row.lease_id = p_lease_id
      AND attempt_row.state = 'running'
      AND attempt_row.lease_until > v_now
    FOR UPDATE;
    IF NOT FOUND OR NOT EXISTS (
        SELECT 1 FROM platform.tenancy_tenant_migration_receipts AS receipt
        WHERE receipt.job_id = p_job_id
          AND receipt.receipt_kind = 'schema_created'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration completion lease is invalid';
    END IF;

    SELECT run.database_origin INTO v_origin
    FROM platform.migration_runs AS run
    WHERE run.run_id = v_run_id AND run.release_id = v_job.release_id
      AND run.environment_id = v_job.environment_id
      AND run.command = 'tenant_migration_executor'
      AND run.state = 'succeeded'
      AND run.migration_manifest_sha256 = v_job.migration_manifest_sha256;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration run is not complete';
    END IF;

    SELECT verification.verified, verification.catalog_sha256,
           verification.requirement_count
    INTO v_verified, v_catalog_sha256, v_requirement_count
    FROM platform.platform_migrator_verify_new_tenant_v1(
        v_job.release_id, v_job.environment_id, v_origin,
        v_operation.tenant_id, v_job.migration_manifest_sha256
    ) AS verification;
    IF NOT v_verified THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration checkpoints are incomplete';
    END IF;

    INSERT INTO platform.tenancy_tenant_migration_receipts (
        receipt_id, job_id, attempt, receipt_kind, migration_run_id,
        receipt_sha256, evidence_sha256, occurred_at
    ) VALUES (
        p_receipt_id, p_job_id, p_attempt, 'migration_completed', v_run_id,
        p_receipt_sha256, p_evidence_sha256, v_now
    );
    UPDATE platform.tenancy_tenant_migration_attempts
    SET state = 'succeeded', finished_at = v_now
    WHERE job_id = p_job_id AND attempt = p_attempt;
    UPDATE platform.tenancy_tenant_migration_jobs
    SET state = 'succeeded', completed_at = v_now, updated_at = v_now
    WHERE job_id = p_job_id;
    UPDATE platform.tenancy_onboarding_operations
    SET state = 'migration_ready', revision = revision + 1, updated_at = v_now
    WHERE operation_id = v_job.operation_id AND state = 'migrating';

    RETURN QUERY SELECT v_catalog_sha256, v_requirement_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_complete_v1(
    uuid, integer, uuid, uuid, uuid, text, text
) FROM PUBLIC;

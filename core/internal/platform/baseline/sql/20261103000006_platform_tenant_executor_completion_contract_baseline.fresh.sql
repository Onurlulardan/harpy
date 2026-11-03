-- +goose Up
-- Forward-only replacement of tenancy_tenant_migration_complete_v3.
-- Goose applies this migration once, inside one transaction. The routine itself
-- is idempotent: a lost completion response can be retried and leaves the same
-- job state, onboarding state, and revision.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_tenant_migration_complete_v3(p_job_id uuid, p_attempt integer, p_profile_instance_id uuid, p_lease_id uuid, p_receipt_id uuid, p_receipt_sha256 text, p_evidence_sha256 text) RETURNS TABLE(catalog_sha256 text, requirement_count integer)
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $fn$
DECLARE
    v_job platform.tenancy_tenant_migration_jobs%ROWTYPE;
    v_operation platform.tenancy_onboarding_operations%ROWTYPE;
    v_run_id uuid;
    v_leased_at timestamptz;
    v_attempt_number bigint;
    v_applied bigint;
    v_max_version bigint;
    v_now timestamptz;
    v_next_state text;
    minimum_schema_version CONSTANT bigint := 20261102000001;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_migrator'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant migration completion authority is denied';
    END IF;
    IF p_receipt_id IS NULL OR p_receipt_sha256 !~ '^[0-9a-f]{64}$'
       OR p_evidence_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenant migration completion receipt is invalid';
    END IF;
    SELECT job.* INTO v_job
    FROM platform.tenancy_tenant_migration_jobs AS job
    WHERE job.job_id = p_job_id
    FOR UPDATE;
    IF NOT FOUND OR v_job.current_attempt <> p_attempt
       OR v_job.state NOT IN ('running', 'succeeded') THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration completion state is invalid';
    END IF;
    SELECT operation.* INTO v_operation
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id = v_job.operation_id
    FOR UPDATE;
    IF NOT FOUND
       OR v_operation.schema_name::text !~ '^t_[a-z2-7]{26}$'
       OR v_operation.operation_kind NOT IN (
           'environment_initialize', 'saas_customer_tenant_provision'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration completion state is invalid';
    END IF;
    v_next_state := CASE
        WHEN v_operation.operation_kind = 'environment_initialize'
         AND v_operation.deployment_mode = 'saas' THEN 'authority_pending'
        ELSE 'migration_ready'
    END;
    IF v_job.state = 'running' AND (
        v_operation.state <> 'migrating' OR EXISTS (
            SELECT 1 FROM platform.tenants WHERE id = v_operation.tenant_id
        )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration completion state is invalid';
    END IF;
    v_now := pg_catalog.clock_timestamp();
    IF v_now <= v_job.updated_at THEN
        v_now := v_job.updated_at + interval '1 microsecond';
    END IF;
    IF v_now <= v_operation.updated_at THEN
        v_now := v_operation.updated_at + interval '1 microsecond';
    END IF;
    IF v_job.state = 'running' THEN
        SELECT attempt_row.migration_run_id, attempt_row.leased_at
        INTO v_run_id, v_leased_at
        FROM platform.tenancy_tenant_migration_attempts AS attempt_row
        WHERE attempt_row.job_id = p_job_id
          AND attempt_row.attempt = p_attempt
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
    ELSE
        SELECT attempt_row.migration_run_id, attempt_row.leased_at
        INTO v_run_id, v_leased_at
        FROM platform.tenancy_tenant_migration_attempts AS attempt_row
        WHERE attempt_row.job_id = p_job_id
          AND attempt_row.attempt = p_attempt
          AND attempt_row.state = 'succeeded';
        IF NOT FOUND OR NOT EXISTS (
            SELECT 1 FROM platform.tenancy_tenant_migration_receipts AS receipt
            WHERE receipt.job_id = p_job_id
              AND receipt.attempt = p_attempt
              AND receipt.receipt_kind = 'migration_completed'
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'tenant migration completion state is invalid';
        END IF;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM platform.environment_identity AS identity
        WHERE identity.singleton
          AND identity.environment_id = v_job.environment_id
    ) OR pg_catalog.to_regnamespace(v_operation.schema_name::text) IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration environment is not current';
    END IF;
    BEGIN
        EXECUTE pg_catalog.format(
            'SELECT COALESCE(MAX(version_id) FILTER (WHERE is_applied AND version_id > 0), 0), COALESCE(count(*) FILTER (WHERE is_applied AND version_id > 0), 0) FROM %I.goose_db_version',
            v_operation.schema_name::text
        ) INTO v_max_version, v_applied;
    EXCEPTION WHEN undefined_table THEN
        v_max_version := 0;
        v_applied := 0;
    END;
    IF v_max_version < minimum_schema_version THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant schema goose history is below the minimum schema version';
    END IF;
    IF v_job.state = 'running' THEN
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
    END IF;
    IF v_operation.state = 'migrating' THEN
        SELECT COALESCE(MAX(attempt.attempt_number), 0) + 1
        INTO v_attempt_number
        FROM platform.tenancy_onboarding_attempts AS attempt
        WHERE attempt.operation_id = v_operation.operation_id;
        INSERT INTO platform.tenancy_onboarding_attempts (
            attempt_id, operation_id, attempt_number, outcome,
            started_at, finished_at
        ) VALUES (
            p_lease_id, v_operation.operation_id, v_attempt_number,
            'completed', COALESCE(v_leased_at, v_now), v_now
        );
        INSERT INTO platform.tenancy_onboarding_receipts (
            receipt_id, operation_id, attempt_id, receipt_kind,
            receipt_sha256, evidence_sha256, occurred_at
        ) VALUES (
            p_receipt_id, v_operation.operation_id, p_lease_id,
            'migration_completed', p_receipt_sha256, p_evidence_sha256, v_now
        );
        UPDATE platform.tenancy_onboarding_operations AS operation
        SET state = v_next_state, revision = operation.revision + 1,
            updated_at = v_now
        WHERE operation.operation_id = v_operation.operation_id
          AND operation.state = 'migrating';
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'tenant migration onboarding transition is invalid';
        END IF;
    ELSIF v_operation.state IS DISTINCT FROM v_next_state THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration onboarding transition is invalid';
    END IF;
    RETURN QUERY SELECT v_job.migration_manifest_sha256, v_applied::integer;
END;
$fn$;
-- +goose StatementEnd

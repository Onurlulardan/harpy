-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_tenant_migration_reopen_unactivated_v1() RETURNS void
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $$
DECLARE
    stuck record;
    triggers_disabled boolean := false;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_migrator'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant migration reopen authority is denied';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM platform.tenancy_tenant_migration_jobs AS job
        JOIN platform.tenancy_onboarding_operations AS operation
          ON operation.operation_id = job.operation_id
        WHERE NOT EXISTS (
            SELECT 1 FROM platform.tenants AS tenant
            WHERE tenant.id = operation.tenant_id
        )
          AND (
              job.state IN (
                  'failed', 'fenced', 'cleanup_required', 'reconcile_required'
              )
              OR (
                  job.state = 'running' AND NOT EXISTS (
                      SELECT 1
                      FROM platform.tenancy_tenant_migration_attempts AS attempt_row
                      JOIN platform.tenancy_tenant_migration_executor_readiness AS ready
                        ON ready.profile_instance_id = attempt_row.profile_instance_id
                      WHERE attempt_row.job_id = job.job_id
                        AND attempt_row.attempt = job.current_attempt
                        AND attempt_row.state = 'running'
                        AND attempt_row.lease_until > pg_catalog.clock_timestamp()
                        AND ready.state = 'active'
                        AND ready.ready_until > pg_catalog.clock_timestamp()
                  )
              )
          )
    ) THEN
        RETURN;
    END IF;

    ALTER TABLE platform.tenancy_tenant_migration_jobs
        DISABLE TRIGGER tenancy_tenant_migration_jobs_guard;
    ALTER TABLE platform.tenancy_tenant_migration_attempts
        DISABLE TRIGGER tenancy_tenant_migration_attempts_guard;
    ALTER TABLE platform.tenancy_tenant_migration_receipts
        DISABLE TRIGGER tenancy_tenant_migration_receipts_guard;
    triggers_disabled := true;

    FOR stuck IN
        SELECT job.job_id, operation.schema_name::text AS schema_name
        FROM platform.tenancy_tenant_migration_jobs AS job
        JOIN platform.tenancy_onboarding_operations AS operation
          ON operation.operation_id = job.operation_id
        WHERE NOT EXISTS (
            SELECT 1 FROM platform.tenants AS tenant
            WHERE tenant.id = operation.tenant_id
        )
          AND (
              job.state IN (
                  'failed', 'fenced', 'cleanup_required', 'reconcile_required'
              )
              OR (
                  job.state = 'running' AND NOT EXISTS (
                      SELECT 1
                      FROM platform.tenancy_tenant_migration_attempts AS attempt_row
                      JOIN platform.tenancy_tenant_migration_executor_readiness AS ready
                        ON ready.profile_instance_id = attempt_row.profile_instance_id
                      WHERE attempt_row.job_id = job.job_id
                        AND attempt_row.attempt = job.current_attempt
                        AND attempt_row.state = 'running'
                        AND attempt_row.lease_until > pg_catalog.clock_timestamp()
                        AND ready.state = 'active'
                        AND ready.ready_until > pg_catalog.clock_timestamp()
                  )
              )
          )
    LOOP
        IF stuck.schema_name !~ '^t_[a-z2-7]{26}$' THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'tenant migration reopen schema is invalid';
        END IF;
        EXECUTE pg_catalog.format(
            'DROP SCHEMA IF EXISTS %I CASCADE', stuck.schema_name
        );
        DELETE FROM platform.tenancy_tenant_migration_receipts
        WHERE job_id = stuck.job_id;
        DELETE FROM platform.tenancy_tenant_migration_attempts
        WHERE job_id = stuck.job_id;
        UPDATE platform.tenancy_tenant_migration_jobs
        SET state = 'pending', current_attempt = 0, error_code = NULL,
            completed_at = NULL, created_at = pg_catalog.clock_timestamp(),
            updated_at = pg_catalog.clock_timestamp()
        WHERE job_id = stuck.job_id;
    END LOOP;

    ALTER TABLE platform.tenancy_tenant_migration_jobs
        ENABLE TRIGGER tenancy_tenant_migration_jobs_guard;
    ALTER TABLE platform.tenancy_tenant_migration_attempts
        ENABLE TRIGGER tenancy_tenant_migration_attempts_guard;
    ALTER TABLE platform.tenancy_tenant_migration_receipts
        ENABLE TRIGGER tenancy_tenant_migration_receipts_guard;
EXCEPTION WHEN OTHERS THEN
    IF triggers_disabled THEN
        ALTER TABLE platform.tenancy_tenant_migration_jobs
            ENABLE TRIGGER tenancy_tenant_migration_jobs_guard;
        ALTER TABLE platform.tenancy_tenant_migration_attempts
            ENABLE TRIGGER tenancy_tenant_migration_attempts_guard;
        ALTER TABLE platform.tenancy_tenant_migration_receipts
            ENABLE TRIGGER tenancy_tenant_migration_receipts_guard;
    END IF;
    RAISE;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
REVOKE ALL ON FUNCTION platform.tenancy_tenant_migration_reopen_unactivated_v1() FROM PUBLIC;
-- +goose StatementEnd

-- +goose StatementBegin
GRANT ALL ON FUNCTION platform.tenancy_tenant_migration_reopen_unactivated_v1() TO harpy_migrator;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_tenant_migration_claim_v1(p_profile_instance_id uuid, p_lease_id uuid, p_migration_run_id uuid, p_expiry_receipt_id uuid, p_lease_seconds integer) RETURNS TABLE(job_id uuid, operation_id uuid, tenant_id uuid, schema_name text, step_plan_sha256 text, origin text, attempt integer, migration_run_id uuid, previous_run_id uuid, lease_id uuid, profile_instance_id uuid, release_id uuid, environment_id uuid, deployment_mode text, runtime_generation bigint, database_generation bigint, migration_manifest_sha256 text, lease_until timestamp with time zone)
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $$
DECLARE
    v_ready platform.tenancy_tenant_migration_executor_readiness%ROWTYPE;
    v_job platform.tenancy_tenant_migration_jobs%ROWTYPE;
    v_previous_run_id uuid;
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_attempt integer;
    v_digest text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_migrator'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='tenant migration claim authority is denied';
    END IF;
    IF p_profile_instance_id IS NULL OR p_lease_id IS NULL OR
       p_migration_run_id IS NULL OR p_expiry_receipt_id IS NULL OR
       p_lease_seconds NOT BETWEEN 30 AND 300 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='tenant migration claim request is invalid';
    END IF;

    SELECT ready.* INTO v_ready
    FROM platform.tenancy_tenant_migration_executor_readiness AS ready
    WHERE ready.profile_instance_id=p_profile_instance_id
      AND ready.state='active' AND ready.ready_until>v_now
    FOR SHARE OF ready;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='tenant migration executor is not claim ready';
    END IF;
    IF EXISTS (
        SELECT 1 FROM platform.tenancy_deployment_policy AS policy
        WHERE policy.environment_id IS DISTINCT FROM v_ready.environment_id
           OR policy.deployment_mode IS DISTINCT FROM v_ready.deployment_mode
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='tenant migration deployment policy conflicts';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM platform.tenancy_deployment_policy) THEN
        RETURN;
    END IF;
    PERFORM platform.tenancy_tenant_migration_reopen_unactivated_v1();
    UPDATE platform.tenancy_tenant_migration_jobs AS exhausted
    SET state='fenced',error_code='attempt_limit_exhausted',
        completed_at=v_now,updated_at=v_now
    WHERE exhausted.state='failed'
      AND exhausted.release_id=v_ready.release_id
      AND exhausted.environment_id=v_ready.environment_id
      AND exhausted.runtime_generation=v_ready.runtime_generation
      AND exhausted.database_generation=v_ready.database_generation
      AND exhausted.migration_manifest_sha256=v_ready.migration_manifest_sha256
      AND (
          SELECT pg_catalog.count(*)
          FROM platform.tenancy_tenant_migration_attempts AS attempt_row
          JOIN platform.tenancy_tenant_migration_executor_readiness AS profile
            ON profile.profile_instance_id=attempt_row.profile_instance_id
          WHERE attempt_row.job_id=exhausted.job_id
            AND profile.release_id=v_ready.release_id
            AND profile.runtime_generation=v_ready.runtime_generation
            AND profile.database_generation=v_ready.database_generation
            AND profile.migration_manifest_sha256=
                v_ready.migration_manifest_sha256
      )>=8;

    SELECT candidate.* INTO v_job
    FROM platform.tenancy_tenant_migration_jobs AS candidate
    WHERE (candidate.state IN ('pending','failed') OR (
        candidate.state='running' AND EXISTS (
            SELECT 1 FROM platform.tenancy_tenant_migration_attempts AS prior
            WHERE prior.job_id=candidate.job_id
              AND prior.attempt=candidate.current_attempt
              AND prior.state='running' AND prior.lease_until<=v_now
        )
    ))
      AND candidate.environment_id=v_ready.environment_id
      AND candidate.deployment_mode=v_ready.deployment_mode
      AND (
        ROW(candidate.release_id,candidate.runtime_generation,
            candidate.database_generation,candidate.migration_manifest_sha256)
        IS DISTINCT FROM ROW(v_ready.release_id,v_ready.runtime_generation,
            v_ready.database_generation,v_ready.migration_manifest_sha256)
        OR (
          SELECT pg_catalog.count(*)
          FROM platform.tenancy_tenant_migration_attempts AS attempt_row
          JOIN platform.tenancy_tenant_migration_executor_readiness AS profile
            ON profile.profile_instance_id=attempt_row.profile_instance_id
          WHERE attempt_row.job_id=candidate.job_id
            AND profile.release_id=v_ready.release_id
            AND profile.runtime_generation=v_ready.runtime_generation
            AND profile.database_generation=v_ready.database_generation
            AND profile.migration_manifest_sha256=
                v_ready.migration_manifest_sha256
        )<8
      )
    ORDER BY candidate.created_at,candidate.job_id
    FOR UPDATE SKIP LOCKED LIMIT 1;
    IF NOT FOUND THEN RETURN; END IF;

    IF v_job.state='running' THEN
        SELECT prior.migration_run_id INTO v_previous_run_id
        FROM platform.tenancy_tenant_migration_attempts AS prior
        WHERE prior.job_id=v_job.job_id AND prior.attempt=v_job.current_attempt
        FOR UPDATE;
        UPDATE platform.tenancy_tenant_migration_attempts
        SET state='failed',finished_at=v_now,error_code='lease_expired'
        WHERE tenancy_tenant_migration_attempts.job_id=v_job.job_id
          AND tenancy_tenant_migration_attempts.attempt=v_job.current_attempt;
        v_digest:=pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
            v_job.job_id::text||':'||v_job.current_attempt::text||
            ':lease_expired','UTF8')),'hex');
        INSERT INTO platform.tenancy_tenant_migration_receipts (
            receipt_id,job_id,attempt,receipt_kind,migration_run_id,
            receipt_sha256,evidence_sha256,occurred_at
        ) VALUES (
            p_expiry_receipt_id,v_job.job_id,v_job.current_attempt,
            'failure',v_previous_run_id,v_digest,v_digest,v_now
        ) ON CONFLICT ON CONSTRAINT
            tenancy_tenant_migration_receipt_attempt_kind_uq DO NOTHING;
    END IF;

    v_attempt:=v_job.current_attempt+1;
    UPDATE platform.tenancy_tenant_migration_jobs
    SET release_id=v_ready.release_id,
        runtime_generation=v_ready.runtime_generation,
        database_generation=v_ready.database_generation,
        migration_manifest_sha256=v_ready.migration_manifest_sha256,
        state='running',current_attempt=v_attempt,error_code=NULL,
        completed_at=NULL,updated_at=v_now
    WHERE tenancy_tenant_migration_jobs.job_id=v_job.job_id;
    INSERT INTO platform.tenancy_tenant_migration_attempts (
        job_id,attempt,migration_run_id,profile_instance_id,lease_id,
        state,leased_at,heartbeat_at,lease_until
    ) VALUES (
        v_job.job_id,v_attempt,p_migration_run_id,p_profile_instance_id,
        p_lease_id,'running',v_now,v_now,
        v_now+pg_catalog.make_interval(secs=>p_lease_seconds)
    );

    RETURN QUERY SELECT
        v_job.job_id,v_job.operation_id,operation.tenant_id,
        operation.schema_name::text,operation.step_plan_sha256,
        v_job.origin,v_attempt,p_migration_run_id,v_previous_run_id,
        p_lease_id,p_profile_instance_id,v_ready.release_id,
        v_ready.environment_id,v_ready.deployment_mode,
        v_ready.runtime_generation,v_ready.database_generation,
        v_ready.migration_manifest_sha256,
        v_now+pg_catalog.make_interval(secs=>p_lease_seconds)
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id=v_job.operation_id;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_tenant_migration_complete_v2(p_job_id uuid, p_attempt integer, p_profile_instance_id uuid, p_lease_id uuid, p_receipt_id uuid, p_receipt_sha256 text, p_evidence_sha256 text) RETURNS TABLE(catalog_sha256 text, requirement_count integer)
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $_$
DECLARE
    v_job platform.tenancy_tenant_migration_jobs%ROWTYPE;
    v_operation platform.tenancy_onboarding_operations%ROWTYPE;
    v_run_id uuid;
    v_applied bigint;
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
    FOR SHARE;
    IF NOT FOUND OR v_operation.state <> 'migrating' OR EXISTS (
           SELECT 1 FROM platform.tenants WHERE id = v_operation.tenant_id
       ) OR v_operation.schema_name::text !~ '^t_[a-z2-7]{26}$' THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration completion state is invalid';
    END IF;
    SELECT attempt_row.migration_run_id INTO v_run_id
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
            'SELECT count(*) FROM %I.goose_db_version WHERE is_applied AND version_id = ANY ($1)',
            v_operation.schema_name::text
        ) INTO v_applied USING ARRAY[20261028000001, 20261102000001]::bigint[];
    EXCEPTION WHEN undefined_table THEN
        v_applied := 0;
    END;
    IF v_applied IS DISTINCT FROM 2 THEN
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
    RETURN QUERY SELECT v_job.migration_manifest_sha256, v_applied::integer;
END;
$_$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $reopen_unactivated$
BEGIN
    IF SESSION_USER = 'harpy_migrator' AND CURRENT_USER = 'harpy_owner' THEN
        PERFORM platform.tenancy_tenant_migration_reopen_unactivated_v1();
    END IF;
END
$reopen_unactivated$;
-- +goose StatementEnd

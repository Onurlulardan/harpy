-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_tenant_migration_ready_register_v1(p_profile_instance_id uuid, p_release_id uuid, p_deployment_mode text, p_runtime_generation bigint, p_migration_manifest_sha256 text, p_ttl_seconds integer) RETURNS TABLE(profile_instance_id uuid, release_id uuid, environment_id uuid, deployment_mode text, runtime_generation bigint, database_generation bigint, migration_manifest_sha256 text, ready_until timestamp with time zone)
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $_$
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

    SELECT identity.environment_id, 1
    INTO v_environment_id, v_database_generation
    FROM platform.environment_identity AS identity
    WHERE identity.singleton = TRUE
      AND identity.database_origin = 'fresh'
      AND identity.database_oid = (
          SELECT database_row.oid::bigint
          FROM pg_catalog.pg_database AS database_row
          WHERE database_row.datname = pg_catalog.current_database()
      );
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration release is not current';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.tenancy_deployment_policy AS policy
        WHERE policy.environment_id IS DISTINCT FROM v_environment_id
           OR policy.deployment_mode IS DISTINCT FROM p_deployment_mode
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration deployment policy conflicts';
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
$_$;
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
CREATE OR REPLACE FUNCTION platform.tenancy_tenant_migration_ready_heartbeat_v1(p_profile_instance_id uuid, p_ttl_seconds integer) RETURNS timestamp with time zone
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $$
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
    WHERE ready.profile_instance_id = p_profile_instance_id
      AND ready.state = 'active'
      AND ready.ready_until > v_now
      AND NOT EXISTS (
          SELECT 1
          FROM platform.tenancy_deployment_policy AS policy
          WHERE policy.environment_id IS DISTINCT FROM ready.environment_id
             OR policy.deployment_mode IS DISTINCT FROM ready.deployment_mode
      )
    RETURNING ready.ready_until INTO v_ready_until;

    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration readiness is stale';
    END IF;
    RETURN v_ready_until;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_tenant_migration_ready_current_v1() RETURNS TABLE(release_id uuid, environment_id uuid, deployment_mode text, runtime_generation bigint, database_generation bigint, migration_manifest_sha256 text, ready_until timestamp with time zone)
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $$
BEGIN
    RETURN QUERY
    SELECT ready.release_id, ready.environment_id, ready.deployment_mode,
           ready.runtime_generation, ready.database_generation,
           ready.migration_manifest_sha256, ready.ready_until
    FROM platform.tenancy_tenant_migration_executor_readiness AS ready
    WHERE ready.state = 'active'
      AND ready.ready_until > pg_catalog.clock_timestamp()
      AND NOT EXISTS (
          SELECT 1
          FROM platform.tenancy_deployment_policy AS policy
          WHERE policy.environment_id IS DISTINCT FROM ready.environment_id
             OR policy.deployment_mode IS DISTINCT FROM ready.deployment_mode
      )
    ORDER BY ready.ready_until DESC, ready.profile_instance_id
    LIMIT 1;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.platform_migrator_verify_new_tenant_v1(p_release_id uuid, p_environment_id uuid, p_database_origin text, p_tenant_id uuid, p_migration_manifest_sha256 text) RETURNS TABLE(verified boolean, catalog_sha256 text, requirement_count integer)
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $$
DECLARE
    v_migration_run_id uuid;
    v_catalog_sha256 text;
    v_requirement_count integer;
BEGIN
    LOCK TABLE platform.tenant_migration_requirement_catalogs,
        platform.tenant_migration_requirements,
        platform.tenant_migration_checkpoints IN SHARE MODE;

    SELECT catalog.migration_run_id, catalog.catalog_sha256,
           catalog.requirement_count
    INTO v_migration_run_id, v_catalog_sha256, v_requirement_count
    FROM platform.tenant_migration_requirement_catalogs AS catalog
    WHERE catalog.release_id = p_release_id
      AND catalog.environment_id = p_environment_id
      AND catalog.database_origin = p_database_origin
      AND catalog.migration_manifest_sha256 = p_migration_manifest_sha256
      AND catalog.contract_version =
          'tenant-migration-requirement-catalog.v2';

    IF NOT FOUND OR v_requirement_count < 1 THEN
        RETURN QUERY SELECT false, NULL::text, 0;
        RETURN;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM platform.tenant_migration_requirements AS requirement
        WHERE requirement.migration_run_id = v_migration_run_id
          AND NOT EXISTS (
              SELECT 1
              FROM platform.tenant_migration_checkpoints AS checkpoint
              WHERE checkpoint.tenant_id = p_tenant_id
                AND checkpoint.owner = requirement.owner
                AND checkpoint.logical_version = requirement.logical_version
                AND checkpoint.database_origin = requirement.database_origin
                AND checkpoint.selected_path = requirement.selected_path
                AND checkpoint.selected_sha256 = requirement.selected_sha256
                AND checkpoint.state = 'succeeded'
          )
    ) OR (
        SELECT pg_catalog.count(*)
        FROM platform.tenant_migration_requirements AS requirement
        WHERE requirement.migration_run_id = v_migration_run_id
    ) IS DISTINCT FROM v_requirement_count::bigint THEN
        RETURN QUERY SELECT false, v_catalog_sha256, v_requirement_count;
        RETURN;
    END IF;

    RETURN QUERY SELECT true, v_catalog_sha256, v_requirement_count;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_tenant_migration_complete_v1(p_job_id uuid, p_attempt integer, p_profile_instance_id uuid, p_lease_id uuid, p_receipt_id uuid, p_receipt_sha256 text, p_evidence_sha256 text) RETURNS TABLE(catalog_sha256 text, requirement_count integer)
    LANGUAGE plpgsql SECURITY DEFINER
    SET search_path TO 'pg_catalog'
    AS $_$
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

    SELECT identity.database_origin INTO v_origin
    FROM platform.environment_identity AS identity
    WHERE identity.singleton
      AND identity.environment_id = v_job.environment_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration environment is not current';
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
$_$;
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
    FOR SHARE;
    IF NOT FOUND OR v_operation.state <> 'migrating' OR EXISTS (
           SELECT 1 FROM platform.tenants WHERE id = v_operation.tenant_id
       ) THEN
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
    SELECT identity.database_origin INTO v_origin
    FROM platform.environment_identity AS identity
    WHERE identity.singleton
      AND identity.environment_id = v_job.environment_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration environment is not current';
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
    RETURN QUERY SELECT v_catalog_sha256, v_requirement_count;
END;
$_$;
-- +goose StatementEnd

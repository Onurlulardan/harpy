-- +goose Up
CREATE TABLE platform.platformops_restore_execution_attempts (
    attempt_id uuid PRIMARY KEY
        CHECK (pg_catalog.substring(attempt_id::text,15,1)='7'),
    operation_id uuid NOT NULL
        REFERENCES platform.platformops_restore_operations(operation_id),
    generation bigint NOT NULL CHECK (generation > 0),
    lease_token uuid NOT NULL UNIQUE
        CHECK (pg_catalog.substring(lease_token::text,15,1)='7'),
    lifecycle_state text NOT NULL CHECK (
        lifecycle_state IN ('active','completed','failed','superseded')
    ),
    claimed_restore_state text NOT NULL CHECK (
        claimed_restore_state IN (
            'plan_ready','draining','restoring','verifying','switched'
        )
    ),
    plan_sha256 text NOT NULL CHECK (plan_sha256 ~ '^[0-9a-f]{64}$'),
    lease_expires_at timestamptz NOT NULL,
    heartbeat_at timestamptz NOT NULL,
    failure_code text CHECK (
        failure_code IS NULL OR failure_code ~ '^[a-z0-9_]{1,64}$'
    ),
    started_at timestamptz NOT NULL,
    completed_at timestamptz,
    UNIQUE (operation_id,generation),
    CHECK ((lifecycle_state IN ('completed','failed','superseded')) =
           (completed_at IS NOT NULL)),
    CHECK ((lifecycle_state='failed') = (failure_code IS NOT NULL))
);
CREATE UNIQUE INDEX platformops_restore_execution_active_key
ON platform.platformops_restore_execution_attempts(operation_id)
WHERE lifecycle_state='active';

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_claim_restore_execution_v2(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid,p_plan_sha256 text
) RETURNS TABLE(
    attempt_id uuid,lease_token uuid,attempt_generation bigint,
    restore_state text,operation_id uuid,recovery_set_id uuid,
    environment_id uuid,target_disposition text,plan_sha256 text,
    plan_bytes bytea,safety_recovery_set_id uuid,
    safety_recovery_set_sha256 text,waiver_reason text,probe_sha256 text,
    destructive_confirmed boolean,manifest_sha256 text,manifest_bytes bytea,
    previous_database_system_id uuid,component_inventory jsonb,
    database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_generation bigint;
    v_now timestamptz := pg_catalog.statement_timestamp();
    v_components jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_restore'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-restore'
       OR pg_catalog.substring(p_attempt_id::text,15,1)<>'7'
       OR pg_catalog.substring(p_lease_token::text,15,1)<>'7'
       OR p_plan_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore execution claim denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'platformops:restore-executor:'||p_operation_id::text,0
        )
    );
    IF EXISTS (
        SELECT 1 FROM platform.platformops_restore_execution_attempts
        WHERE attempt_id=p_attempt_id
          AND (operation_id<>p_operation_id OR lease_token<>p_lease_token
               OR plan_sha256<>p_plan_sha256
               OR lifecycle_state<>'active' OR lease_expires_at<=v_now)
    ) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='restore execution claim conflicts';
    END IF;
    PERFORM 1 FROM platform.platformops_restore_operations AS restore
    JOIN platform.platformops_recovery_sets AS recovery
      ON recovery.recovery_set_id=restore.recovery_set_id
    WHERE restore.operation_id=p_operation_id
      AND restore.plan_sha256=p_plan_sha256
      AND restore.state IN (
          'plan_ready','draining','restoring','verifying','switched'
      )
      AND recovery.state='protected'
      AND recovery.manifest_sha256 IS NOT NULL
      AND recovery.manifest_bytes IS NOT NULL
    FOR UPDATE OF restore;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='restore operation is not claimable';
    END IF;
    SELECT pg_catalog.jsonb_object_agg(
        component.component_kind,
        pg_catalog.jsonb_build_object(
            'object_key',component.object_key,
            'object_version_id',component.object_version_id,
            'ciphertext_sha256',component.ciphertext_sha256,
            'ciphertext_bytes',component.ciphertext_bytes
        )
    ) INTO v_components
    FROM platform.platformops_recovery_components AS component
    JOIN platform.platformops_restore_operations AS restore
      ON restore.recovery_set_id=component.recovery_set_id
    WHERE restore.operation_id=p_operation_id;
    IF pg_catalog.jsonb_object_length(v_components)<>5
       OR NOT v_components ?& ARRAY[
           'database_archive','documents_inventory','key_custody',
           'owner_anchor','protected_authority'
       ] THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='restore component inventory is incomplete';
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM platform.platformops_restore_execution_attempts
        WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
          AND lifecycle_state='active' AND lease_expires_at>v_now
    ) THEN
        UPDATE platform.platformops_restore_execution_attempts
        SET lifecycle_state='superseded',completed_at=v_now
        WHERE operation_id=p_operation_id AND lifecycle_state='active'
          AND lease_expires_at<=v_now;
        IF EXISTS (
            SELECT 1 FROM platform.platformops_restore_execution_attempts
            WHERE operation_id=p_operation_id AND lifecycle_state='active'
        ) THEN
            RAISE EXCEPTION USING ERRCODE='55P03',
                MESSAGE='restore execution is leased';
        END IF;
        SELECT COALESCE(pg_catalog.max(generation),0)+1 INTO v_generation
        FROM platform.platformops_restore_execution_attempts
        WHERE operation_id=p_operation_id;
        INSERT INTO platform.platformops_restore_execution_attempts(
            attempt_id,operation_id,generation,lease_token,lifecycle_state,
            claimed_restore_state,plan_sha256,lease_expires_at,
            heartbeat_at,started_at
        )
        SELECT p_attempt_id,p_operation_id,v_generation,p_lease_token,'active',
               state,p_plan_sha256,v_now+interval '2 minutes',v_now,v_now
        FROM platform.platformops_restore_operations
        WHERE operation_id=p_operation_id;
    END IF;
    RETURN QUERY
    SELECT attempt.attempt_id,attempt.lease_token,attempt.generation,
           restore.state,restore.operation_id,restore.recovery_set_id,
           restore.environment_id,restore.target_disposition,
           restore.plan_sha256,restore.plan_bytes,
           restore.safety_recovery_set_id,restore.safety_recovery_set_sha256,
           restore.waiver_reason,restore.probe_sha256,
           restore.destructive_confirmed,recovery.manifest_sha256,
           recovery.manifest_bytes,recovery.database_system_id,
           v_components,v_now
    FROM platform.platformops_restore_execution_attempts AS attempt
    JOIN platform.platformops_restore_operations AS restore
      ON restore.operation_id=attempt.operation_id
    JOIN platform.platformops_recovery_sets AS recovery
      ON recovery.recovery_set_id=restore.recovery_set_id
    WHERE attempt.attempt_id=p_attempt_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_claim_restore_execution_v2(
    uuid,uuid,uuid,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.platformops_claim_restore_v1(
    uuid,timestamptz
) FROM forgecrud_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_restore_execution_v2(
    uuid,uuid,uuid,text
) TO forgecrud_restore;

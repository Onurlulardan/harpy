-- +goose Up
ALTER TABLE platform.platformops_restore_execution_attempts
    ADD COLUMN control_database_locator_sha256 text
        CHECK (control_database_locator_sha256 IS NULL OR
               control_database_locator_sha256 ~ '^[0-9a-f]{64}$'),
    ADD COLUMN scratch_database_locator_sha256 text
        CHECK (scratch_database_locator_sha256 IS NULL OR
               scratch_database_locator_sha256 ~ '^[0-9a-f]{64}$'),
    ADD COLUMN switch_ack_sha256 text
        CHECK (switch_ack_sha256 IS NULL OR
               switch_ack_sha256 ~ '^[0-9a-f]{64}$');

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_claim_restore_execution_v4(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid,p_plan_sha256 text
) RETURNS TABLE(
    attempt_id uuid,lease_token uuid,attempt_generation bigint,
    restore_state text,operation_id uuid,recovery_set_id uuid,
    environment_id uuid,target_disposition text,plan_sha256 text,
    plan_bytes bytea,safety_recovery_set_id uuid,
    safety_recovery_set_sha256 text,waiver_reason text,probe_sha256 text,
    destructive_confirmed boolean,manifest_sha256 text,manifest_bytes bytea,
    previous_database_system_id uuid,component_inventory jsonb,
    control_database_locator_sha256 text,
    scratch_database_locator_sha256 text,drain_ack_sha256 text,
    switch_ack_sha256 text,database_now timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
RETURN QUERY
WITH claimed AS MATERIALIZED (
    SELECT * FROM platform.platformops_claim_restore_execution_v3(
        p_operation_id,p_attempt_id,p_lease_token,p_plan_sha256
    )
)
SELECT claimed.attempt_id,claimed.lease_token,claimed.attempt_generation,
       claimed.restore_state,claimed.operation_id,claimed.recovery_set_id,
       claimed.environment_id,claimed.target_disposition,
       claimed.plan_sha256,claimed.plan_bytes,
       claimed.safety_recovery_set_id,claimed.safety_recovery_set_sha256,
       claimed.waiver_reason,claimed.probe_sha256,
       claimed.destructive_confirmed,claimed.manifest_sha256,
       claimed.manifest_bytes,claimed.previous_database_system_id,
       claimed.component_inventory,
       COALESCE(attempt.control_database_locator_sha256,''),
       COALESCE(attempt.scratch_database_locator_sha256,''),
       COALESCE(attempt.drain_ack_sha256,''),
       COALESCE(attempt.switch_ack_sha256,''),claimed.database_now
FROM claimed
JOIN platform.platformops_restore_execution_attempts AS attempt
  ON attempt.attempt_id=claimed.attempt_id
 AND attempt.lease_token=claimed.lease_token;
END;
$function$;

CREATE FUNCTION platform.platformops_start_restore_apply_v2(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid,
    p_drain_ack_sha256 text,p_control_database_locator_sha256 text,
    p_scratch_database_locator_sha256 text
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_changed boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR p_drain_ack_sha256 !~ '^[0-9a-f]{64}$'
       OR p_control_database_locator_sha256 !~ '^[0-9a-f]{64}$'
       OR p_scratch_database_locator_sha256 !~ '^[0-9a-f]{64}$'
       OR p_control_database_locator_sha256=
          p_scratch_database_locator_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore database locator binding denied';
    END IF;
    SELECT platform.platformops_start_restore_apply_v1(
        p_operation_id,p_attempt_id,p_lease_token,p_drain_ack_sha256
    ) INTO v_changed;
    IF v_changed THEN
        UPDATE platform.platformops_restore_execution_attempts
        SET control_database_locator_sha256=
                p_control_database_locator_sha256,
            scratch_database_locator_sha256=
                p_scratch_database_locator_sha256
        WHERE attempt_id=p_attempt_id AND operation_id=p_operation_id
          AND lease_token=p_lease_token
          AND lifecycle_state='active'
          AND drain_ack_sha256=p_drain_ack_sha256
          AND control_database_locator_sha256 IS NULL
          AND scratch_database_locator_sha256 IS NULL;
        RETURN FOUND;
    END IF;
    RETURN EXISTS (
        SELECT 1
        FROM platform.platformops_restore_execution_attempts AS attempt
        JOIN platform.platformops_restore_operations AS restore
          ON restore.operation_id=attempt.operation_id
        WHERE attempt.attempt_id=p_attempt_id
          AND attempt.operation_id=p_operation_id
          AND attempt.lease_token=p_lease_token
          AND attempt.lifecycle_state='active'
          AND attempt.claimed_restore_state='restoring'
          AND restore.state='restoring'
          AND attempt.drain_ack_sha256=p_drain_ack_sha256
          AND attempt.control_database_locator_sha256=
              p_control_database_locator_sha256
          AND attempt.scratch_database_locator_sha256=
              p_scratch_database_locator_sha256
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_claim_restore_execution_v4(
    uuid,uuid,uuid,text
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_start_restore_apply_v2(
    uuid,uuid,uuid,text,text,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.platformops_claim_restore_execution_v3(
    uuid,uuid,uuid,text
) FROM harpy_restore;
REVOKE EXECUTE ON FUNCTION platform.platformops_start_restore_apply_v1(
    uuid,uuid,uuid,text
) FROM harpy_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_claim_restore_execution_v4(
    uuid,uuid,uuid,text
) TO harpy_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_start_restore_apply_v2(
    uuid,uuid,uuid,text,text,text
) TO harpy_restore;

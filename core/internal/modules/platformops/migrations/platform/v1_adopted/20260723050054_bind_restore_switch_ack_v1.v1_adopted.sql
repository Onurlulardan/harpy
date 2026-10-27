-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_mark_restore_switched_v2(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid,
    p_object_merkle_sha256 text,p_locator_remap_sha256 text,
    p_safety_evidence_sha256 text,p_switch_ack_sha256 text,
    p_previous_database_system_id uuid,p_new_database_system_id uuid,
    p_control_database_locator_sha256 text,
    p_scratch_database_locator_sha256 text
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_changed boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR p_switch_ack_sha256 !~ '^[0-9a-f]{64}$'
       OR p_control_database_locator_sha256 !~ '^[0-9a-f]{64}$'
       OR p_scratch_database_locator_sha256 !~ '^[0-9a-f]{64}$'
       OR p_control_database_locator_sha256=
          p_scratch_database_locator_sha256
       OR p_previous_database_system_id IS NULL
       OR p_new_database_system_id IS NULL
       OR p_previous_database_system_id=p_new_database_system_id THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore switch authority binding denied';
    END IF;
    PERFORM 1
    FROM platform.platformops_restore_execution_attempts AS attempt
    JOIN platform.platformops_restore_operations AS restore
      ON restore.operation_id=attempt.operation_id
    JOIN platform.platformops_recovery_sets AS recovery
      ON recovery.recovery_set_id=restore.recovery_set_id
    WHERE attempt.attempt_id=p_attempt_id
      AND attempt.operation_id=p_operation_id
      AND attempt.lease_token=p_lease_token
      AND attempt.lifecycle_state='active'
      AND attempt.claimed_restore_state='verifying'
      AND restore.state='verifying'
      AND recovery.database_system_id=p_previous_database_system_id
      AND attempt.restored_database_system_id=p_new_database_system_id
      AND attempt.control_database_locator_sha256=
          p_control_database_locator_sha256
      AND attempt.scratch_database_locator_sha256=
          p_scratch_database_locator_sha256
      AND attempt.switch_ack_sha256 IS NULL
    FOR UPDATE OF attempt,restore;
    IF NOT FOUND THEN
        RETURN EXISTS (
            SELECT 1
            FROM platform.platformops_restore_execution_attempts AS attempt
            JOIN platform.platformops_restore_operations AS restore
              ON restore.operation_id=attempt.operation_id
            WHERE attempt.attempt_id=p_attempt_id
              AND attempt.operation_id=p_operation_id
              AND attempt.lease_token=p_lease_token
              AND attempt.lifecycle_state='active'
              AND attempt.claimed_restore_state='switched'
              AND restore.state='switched'
              AND attempt.switch_ack_sha256=p_switch_ack_sha256
              AND attempt.restored_database_system_id=p_new_database_system_id
              AND attempt.control_database_locator_sha256=
                  p_control_database_locator_sha256
              AND attempt.scratch_database_locator_sha256=
                  p_scratch_database_locator_sha256
        );
    END IF;
    SELECT platform.platformops_mark_restore_switched_v1(
        p_operation_id,p_attempt_id,p_lease_token,
        p_object_merkle_sha256,p_locator_remap_sha256,
        p_safety_evidence_sha256
    ) INTO v_changed;
    IF NOT v_changed THEN RETURN false; END IF;
    UPDATE platform.platformops_restore_execution_attempts
    SET switch_ack_sha256=p_switch_ack_sha256
    WHERE attempt_id=p_attempt_id AND operation_id=p_operation_id
      AND lease_token=p_lease_token
      AND lifecycle_state='active'
      AND claimed_restore_state='switched'
      AND switch_ack_sha256 IS NULL;
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_mark_restore_switched_v2(
    uuid,uuid,uuid,text,text,text,text,uuid,uuid,text,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.platformops_mark_restore_switched_v1(
    uuid,uuid,uuid,text,text,text
) FROM harpy_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_mark_restore_switched_v2(
    uuid,uuid,uuid,text,text,text,text,uuid,uuid,text,text
) TO harpy_restore;

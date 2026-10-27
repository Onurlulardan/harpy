-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_complete_restore_execution_v2(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid,
    p_receipt_sha256 text,p_receipt_bytes bytea
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.statement_timestamp();
    v_receipt jsonb;
    v_attempt platform.platformops_restore_execution_attempts%ROWTYPE;
    v_restore platform.platformops_restore_operations%ROWTYPE;
    v_recovery platform.platformops_recovery_sets%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_restore'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-restore'
       OR p_receipt_sha256 !~ '^[0-9a-f]{64}$'
       OR pg_catalog.octet_length(p_receipt_bytes) NOT BETWEEN 1 AND 2097152
       OR p_receipt_sha256<>pg_catalog.encode(
           pg_catalog.sha256(p_receipt_bytes),'hex'
       ) THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore completion v2 denied';
    END IF;
    SELECT * INTO v_attempt
    FROM platform.platformops_restore_execution_attempts
    WHERE attempt_id=p_attempt_id AND operation_id=p_operation_id
      AND lease_token=p_lease_token FOR UPDATE;
    SELECT * INTO v_restore
    FROM platform.platformops_restore_operations
    WHERE operation_id=p_operation_id FOR UPDATE;
    SELECT * INTO v_recovery
    FROM platform.platformops_recovery_sets
    WHERE recovery_set_id=v_restore.recovery_set_id;
    IF v_attempt.lifecycle_state='completed'
       AND v_restore.state='completed' THEN
        RETURN (v_restore.receipt_sha256,v_restore.receipt_bytes,
                v_restore.new_database_system_id)
            IS NOT DISTINCT FROM
               (p_receipt_sha256,p_receipt_bytes,
                v_attempt.restored_database_system_id);
    END IF;
    IF v_attempt.lifecycle_state<>'active'
       OR v_attempt.claimed_restore_state<>'switched'
       OR v_attempt.lease_expires_at<=v_now
       OR v_attempt.drain_ack_sha256 IS NULL
       OR v_attempt.switch_ack_sha256 IS NULL
       OR v_attempt.control_database_locator_sha256 IS NULL
       OR v_attempt.scratch_database_locator_sha256 IS NULL
       OR v_restore.state<>'switched' THEN
        RETURN false;
    END IF;
    v_receipt := pg_catalog.convert_from(p_receipt_bytes,'UTF8')::jsonb;
    IF v_receipt->>'schema' IS DISTINCT FROM
          'forgecrud-recovery-artifact.v1'
       OR v_receipt->>'purpose' IS DISTINCT FROM 'restore_receipt'
       OR v_receipt->>'environment_id' IS DISTINCT FROM
          v_restore.environment_id::text
       OR v_receipt->>'operation_id' IS DISTINCT FROM p_operation_id::text
       OR v_receipt->>'artifact_id' IS DISTINCT FROM p_operation_id::text
       OR (v_receipt->>'artifact_sequence')::bigint IS DISTINCT FROM
          v_recovery.sequence
       OR (v_receipt->>'key_generation')::bigint IS DISTINCT FROM
          v_recovery.recovery_signing_generation
       OR v_receipt->>'algorithm' IS DISTINCT FROM 'Ed25519'
       OR v_receipt->>'payload_sha256' !~ '^[0-9a-f]{64}$'
       OR pg_catalog.octet_length(v_receipt->>'signature')<1
       OR (v_receipt->>'issued_at')::timestamptz>v_now
       OR (v_receipt->>'expires_at')::timestamptz<=v_now
       OR v_receipt#>>'{payload,schema}' IS DISTINCT FROM
          'restore-receipt-payload.v2'
       OR v_receipt#>>'{payload,operation_id}' IS DISTINCT FROM
          p_operation_id::text
       OR v_receipt#>>'{payload,recovery_set_id}' IS DISTINCT FROM
          v_restore.recovery_set_id::text
       OR v_receipt#>>'{payload,plan_sha256}' IS DISTINCT FROM
          v_restore.plan_sha256
       OR v_receipt#>>'{payload,manifest_sha256}' IS DISTINCT FROM
          v_recovery.manifest_sha256
       OR v_receipt#>>'{payload,identity,environment_id}' IS DISTINCT FROM
          v_restore.environment_id::text
       OR v_receipt#>>'{payload,previous_database_system_id}' IS DISTINCT FROM
          v_recovery.database_system_id::text
       OR v_receipt#>>'{payload,new_database_system_id}' IS DISTINCT FROM
          v_attempt.restored_database_system_id::text
       OR v_receipt#>>'{payload,previous_database_locator_sha256}'
          IS DISTINCT FROM v_attempt.control_database_locator_sha256
       OR v_receipt#>>'{payload,new_database_locator_sha256}'
          IS DISTINCT FROM v_attempt.scratch_database_locator_sha256
       OR v_receipt#>>'{payload,drain_ack_sha256}' IS DISTINCT FROM
          v_attempt.drain_ack_sha256
       OR v_receipt#>>'{payload,switch_ack_sha256}' IS DISTINCT FROM
          v_attempt.switch_ack_sha256
       OR v_receipt#>>'{payload,object_merkle_sha256}' IS DISTINCT FROM
          v_attempt.object_merkle_sha256
       OR v_receipt#>>'{payload,locator_remap_sha256}' IS DISTINCT FROM
          v_attempt.locator_remap_sha256
       OR v_receipt#>>'{payload,safety_evidence_sha256}' IS DISTINCT FROM
          v_attempt.safety_evidence_sha256
       OR (v_receipt#>>'{payload,seeded}')::boolean
       OR (v_receipt#>>'{payload,owner_changed}')::boolean
       OR (v_receipt#>>'{payload,tenant_replaced}')::boolean
       OR v_receipt#>>'{payload,switch_outcome}' IS DISTINCT FROM
          'switched_same_environment' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore receipt v2 evidence denied';
    END IF;
    UPDATE platform.platformops_restore_operations
    SET state='completed',receipt_sha256=p_receipt_sha256,
        receipt_bytes=p_receipt_bytes,
        new_database_system_id=v_attempt.restored_database_system_id,
        completed_at=v_now,updated_at=v_now
    WHERE operation_id=p_operation_id AND state='switched';
    IF NOT FOUND THEN RETURN false; END IF;
    UPDATE platform.platformops_restore_execution_attempts
    SET lifecycle_state='completed',heartbeat_at=v_now,
        lease_expires_at=v_now,completed_at=v_now
    WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
      AND lifecycle_state='active';
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_complete_restore_execution_v2(
    uuid,uuid,uuid,text,bytea
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.platformops_complete_restore_execution_v1(
    uuid,uuid,uuid,text,bytea
) FROM forgecrud_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_complete_restore_execution_v2(
    uuid,uuid,uuid,text,bytea
) TO forgecrud_restore;

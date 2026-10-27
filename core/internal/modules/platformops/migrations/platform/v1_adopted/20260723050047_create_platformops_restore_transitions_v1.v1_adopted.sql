-- +goose Up
ALTER TABLE platform.platformops_restore_execution_attempts
    ADD COLUMN drain_ack_sha256 text
        CHECK (drain_ack_sha256 IS NULL OR drain_ack_sha256 ~ '^[0-9a-f]{64}$'),
    ADD COLUMN archive_readback_sha256 text
        CHECK (archive_readback_sha256 IS NULL OR
               archive_readback_sha256 ~ '^[0-9a-f]{64}$'),
    ADD COLUMN restored_database_system_id uuid,
    ADD COLUMN object_merkle_sha256 text
        CHECK (object_merkle_sha256 IS NULL OR
               object_merkle_sha256 ~ '^[0-9a-f]{64}$'),
    ADD COLUMN locator_remap_sha256 text
        CHECK (locator_remap_sha256 IS NULL OR
               locator_remap_sha256 ~ '^[0-9a-f]{64}$'),
    ADD COLUMN safety_evidence_sha256 text
        CHECK (safety_evidence_sha256 IS NULL OR
               safety_evidence_sha256 ~ '^[0-9a-f]{64}$');

-- +goose StatementBegin
CREATE FUNCTION platform.platformops_start_restore_drain_v1(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'harpy-v2:core-database-restore' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore drain transition denied';
    END IF;
    UPDATE platform.platformops_restore_operations AS restore
    SET state='draining',updated_at=v_now
    FROM platform.platformops_restore_execution_attempts AS attempt
    WHERE restore.operation_id=p_operation_id AND restore.state='plan_ready'
      AND attempt.attempt_id=p_attempt_id
      AND attempt.operation_id=restore.operation_id
      AND attempt.lease_token=p_lease_token
      AND attempt.lifecycle_state='active' AND attempt.lease_expires_at>v_now;
    IF NOT FOUND THEN RETURN false; END IF;
    UPDATE platform.platformops_restore_execution_attempts
    SET claimed_restore_state='draining',heartbeat_at=v_now,
        lease_expires_at=v_now+interval '2 minutes'
    WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
      AND lifecycle_state='active';
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_start_restore_apply_v1(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid,
    p_drain_ack_sha256 text
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'harpy-v2:core-database-restore'
       OR p_drain_ack_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore apply transition denied';
    END IF;
    UPDATE platform.platformops_restore_operations AS restore
    SET state='restoring',updated_at=v_now
    FROM platform.platformops_restore_execution_attempts AS attempt
    WHERE restore.operation_id=p_operation_id AND restore.state='draining'
      AND attempt.attempt_id=p_attempt_id
      AND attempt.operation_id=restore.operation_id
      AND attempt.lease_token=p_lease_token
      AND attempt.lifecycle_state='active' AND attempt.lease_expires_at>v_now
      AND (attempt.drain_ack_sha256 IS NULL
           OR attempt.drain_ack_sha256=p_drain_ack_sha256);
    IF NOT FOUND THEN RETURN false; END IF;
    UPDATE platform.platformops_restore_execution_attempts
    SET claimed_restore_state='restoring',
        drain_ack_sha256=p_drain_ack_sha256,heartbeat_at=v_now,
        lease_expires_at=v_now+interval '2 minutes'
    WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
      AND lifecycle_state='active';
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_start_restore_verify_v1(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid,
    p_archive_readback_sha256 text,p_new_database_system_id uuid
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'harpy-v2:core-database-restore'
       OR p_archive_readback_sha256 !~ '^[0-9a-f]{64}$'
       OR p_new_database_system_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore verify transition denied';
    END IF;
    UPDATE platform.platformops_restore_operations AS restore
    SET state='verifying',updated_at=v_now
    FROM platform.platformops_restore_execution_attempts AS attempt
    WHERE restore.operation_id=p_operation_id AND restore.state='restoring'
      AND attempt.attempt_id=p_attempt_id
      AND attempt.operation_id=restore.operation_id
      AND attempt.lease_token=p_lease_token
      AND attempt.lifecycle_state='active' AND attempt.lease_expires_at>v_now
      AND attempt.drain_ack_sha256 IS NOT NULL
      AND (attempt.archive_readback_sha256 IS NULL
           OR (attempt.archive_readback_sha256=p_archive_readback_sha256
               AND attempt.restored_database_system_id=p_new_database_system_id));
    IF NOT FOUND THEN RETURN false; END IF;
    UPDATE platform.platformops_restore_execution_attempts
    SET claimed_restore_state='verifying',
        archive_readback_sha256=p_archive_readback_sha256,
        restored_database_system_id=p_new_database_system_id,
        heartbeat_at=v_now,lease_expires_at=v_now+interval '2 minutes'
    WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
      AND lifecycle_state='active';
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_start_restore_drain_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_start_restore_apply_v1(
    uuid,uuid,uuid,text
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_start_restore_verify_v1(
    uuid,uuid,uuid,text,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_start_restore_drain_v1(
    uuid,uuid,uuid
) TO harpy_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_start_restore_apply_v1(
    uuid,uuid,uuid,text
) TO harpy_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_start_restore_verify_v1(
    uuid,uuid,uuid,text,uuid
) TO harpy_restore;

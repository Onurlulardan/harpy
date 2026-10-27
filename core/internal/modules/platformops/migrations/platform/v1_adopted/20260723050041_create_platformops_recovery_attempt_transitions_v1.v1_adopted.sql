-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_recovery_attempt_transition_v1(
    p_recovery_set_id uuid,p_attempt_id uuid,p_lease_token uuid,
    p_expected_state text,p_next_state text
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM platform.platformops_recovery_execution_attempts AS attempt
        WHERE attempt.attempt_id=p_attempt_id
          AND attempt.recovery_set_id=p_recovery_set_id
          AND attempt.lease_token=p_lease_token
          AND attempt.lifecycle_state='active'
          AND attempt.lease_expires_at>v_now
    ) THEN RETURN false; END IF;
    UPDATE platform.platformops_recovery_sets
    SET state=p_next_state,updated_at=v_now
    WHERE recovery_set_id=p_recovery_set_id AND state=p_expected_state;
    IF NOT FOUND THEN RETURN false; END IF;
    UPDATE platform.platformops_recovery_execution_attempts
    SET claimed_recovery_state=p_next_state,heartbeat_at=v_now,
        lease_expires_at=v_now+interval '2 minutes'
    WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
      AND lifecycle_state='active';
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_start_recovery_snapshot_v1(
    p_recovery_set_id uuid,p_attempt_id uuid,p_lease_token uuid
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'harpy-v2:core-database-backup-executor' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery snapshot transition denied';
    END IF;
    RETURN platform.platformops_recovery_attempt_transition_v1(
        p_recovery_set_id,p_attempt_id,p_lease_token,'holding','snapshotting'
    );
END;
$function$;

CREATE FUNCTION platform.platformops_start_recovery_copy_v1(
    p_recovery_set_id uuid,p_attempt_id uuid,p_lease_token uuid
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'harpy-v2:core-database-backup-executor' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery copy transition denied';
    END IF;
    RETURN platform.platformops_recovery_attempt_transition_v1(
        p_recovery_set_id,p_attempt_id,p_lease_token,'snapshotting','copying'
    );
END;
$function$;

CREATE FUNCTION platform.platformops_start_recovery_verify_v1(
    p_recovery_set_id uuid,p_attempt_id uuid,p_lease_token uuid
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'harpy-v2:core-database-backup-executor' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery verify transition denied';
    END IF;
    RETURN platform.platformops_recovery_attempt_transition_v1(
        p_recovery_set_id,p_attempt_id,p_lease_token,'copying','verifying'
    );
END;
$function$;

CREATE FUNCTION platform.platformops_attest_documents_hold_v1(
    p_recovery_set_id uuid,p_attempt_id uuid,p_lease_token uuid,
    p_hold_sha256 text,p_inventory_count bigint
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'harpy-v2:core-database-backup-executor'
       OR p_hold_sha256 !~ '^[0-9a-f]{64}$' OR p_inventory_count<0 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='documents hold attestation denied';
    END IF;
    UPDATE platform.platformops_recovery_execution_attempts AS attempt
    SET documents_hold_sha256=p_hold_sha256,
        documents_inventory_count=p_inventory_count,
        heartbeat_at=v_now,lease_expires_at=v_now+interval '2 minutes'
    FROM platform.platformops_recovery_sets AS recovery
    WHERE attempt.attempt_id=p_attempt_id
      AND attempt.recovery_set_id=p_recovery_set_id
      AND attempt.lease_token=p_lease_token
      AND attempt.lifecycle_state='active' AND attempt.lease_expires_at>v_now
      AND recovery.recovery_set_id=attempt.recovery_set_id
      AND recovery.state IN ('copying','verifying')
      AND (attempt.documents_hold_sha256 IS NULL
           OR (attempt.documents_hold_sha256=p_hold_sha256
               AND attempt.documents_inventory_count=p_inventory_count));
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_fail_recovery_execution_v1(
    p_recovery_set_id uuid,p_attempt_id uuid,p_lease_token uuid,p_failure_code text
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_backup'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'harpy-v2:core-database-backup-executor'
       OR p_failure_code !~ '^[a-z0-9_]{1,64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='recovery execution failure denied';
    END IF;
    UPDATE platform.platformops_recovery_sets AS recovery
    SET state='failed',failure_code=p_failure_code,updated_at=v_now
    FROM platform.platformops_recovery_execution_attempts AS attempt
    WHERE recovery.recovery_set_id=p_recovery_set_id
      AND recovery.state IN ('holding','snapshotting','copying','verifying')
      AND attempt.attempt_id=p_attempt_id
      AND attempt.recovery_set_id=recovery.recovery_set_id
      AND attempt.lease_token=p_lease_token
      AND attempt.lifecycle_state='active' AND attempt.lease_expires_at>v_now;
    IF NOT FOUND THEN RETURN false; END IF;
    UPDATE platform.platformops_recovery_execution_attempts
    SET lifecycle_state='failed',failure_code=p_failure_code,completed_at=v_now
    WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
      AND lifecycle_state='active';
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_recovery_attempt_transition_v1(uuid,uuid,uuid,text,text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_start_recovery_snapshot_v1(uuid,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_start_recovery_copy_v1(uuid,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_start_recovery_verify_v1(uuid,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_attest_documents_hold_v1(uuid,uuid,uuid,text,bigint) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_fail_recovery_execution_v1(uuid,uuid,uuid,text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_start_recovery_snapshot_v1(uuid,uuid,uuid) TO harpy_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_start_recovery_copy_v1(uuid,uuid,uuid) TO harpy_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_start_recovery_verify_v1(uuid,uuid,uuid) TO harpy_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_attest_documents_hold_v1(uuid,uuid,uuid,text,bigint) TO harpy_backup_reader;
GRANT EXECUTE ON FUNCTION platform.platformops_fail_recovery_execution_v1(uuid,uuid,uuid,text) TO harpy_backup_reader;

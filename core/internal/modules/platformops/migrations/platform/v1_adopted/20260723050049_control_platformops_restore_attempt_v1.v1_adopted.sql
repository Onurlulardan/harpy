-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_heartbeat_restore_execution_v1(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-restore' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore heartbeat denied';
    END IF;
    UPDATE platform.platformops_restore_execution_attempts
    SET heartbeat_at=v_now,lease_expires_at=v_now+interval '2 minutes'
    WHERE attempt_id=p_attempt_id AND operation_id=p_operation_id
      AND lease_token=p_lease_token
      AND lifecycle_state='active' AND lease_expires_at>v_now;
    RETURN FOUND;
END;
$function$;

CREATE FUNCTION platform.platformops_fail_restore_execution_v1(
    p_operation_id uuid,p_attempt_id uuid,p_lease_token uuid,
    p_failure_code text,p_fenced boolean
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_restore'
       OR pg_catalog.current_setting('application_name')
          IS DISTINCT FROM 'forgecrud-v2:core-database-restore'
       OR p_failure_code !~ '^[a-z0-9_]{1,64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='restore failure transition denied';
    END IF;
    UPDATE platform.platformops_restore_operations AS restore
    SET state=CASE WHEN p_fenced THEN 'fenced' ELSE 'failed' END,
        updated_at=v_now
    FROM platform.platformops_restore_execution_attempts AS attempt
    WHERE restore.operation_id=p_operation_id
      AND restore.state IN (
          'plan_ready','draining','restoring','verifying','switched'
      )
      AND attempt.attempt_id=p_attempt_id
      AND attempt.operation_id=restore.operation_id
      AND attempt.lease_token=p_lease_token
      AND attempt.lifecycle_state='active' AND attempt.lease_expires_at>v_now;
    IF NOT FOUND THEN RETURN false; END IF;
    UPDATE platform.platformops_restore_execution_attempts
    SET lifecycle_state='failed',failure_code=p_failure_code,
        heartbeat_at=v_now,lease_expires_at=v_now,completed_at=v_now
    WHERE attempt_id=p_attempt_id AND lease_token=p_lease_token
      AND lifecycle_state='active';
    RETURN FOUND;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_heartbeat_restore_execution_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformops_fail_restore_execution_v1(
    uuid,uuid,uuid,text,boolean
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_heartbeat_restore_execution_v1(
    uuid,uuid,uuid
) TO harpy_restore;
GRANT EXECUTE ON FUNCTION platform.platformops_fail_restore_execution_v1(
    uuid,uuid,uuid,text,boolean
) TO harpy_restore;

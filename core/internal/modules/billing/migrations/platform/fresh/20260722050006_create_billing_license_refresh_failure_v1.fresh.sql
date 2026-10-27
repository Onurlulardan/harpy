-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_license_refresh_failure_v1(
    p_operation_id uuid, p_lease_token uuid, p_failure_kind text
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.billing_license_refresh_operations%ROWTYPE;
    v_attempt_state text;
    v_operation_state text;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF p_failure_kind NOT IN (
        'source_unavailable', 'verification_failed',
        'sequence_conflict', 'sequence_rollback'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing license refresh failure kind is invalid';
    END IF;
    SELECT operation.* INTO v_operation
    FROM platform.billing_license_refresh_operations AS operation
    WHERE operation.operation_id = p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR v_operation.state <> 'running'
       OR v_operation.lease_token IS DISTINCT FROM p_lease_token THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'billing license refresh failure lease is stale';
    END IF;
    v_attempt_state := CASE p_failure_kind
        WHEN 'source_unavailable' THEN 'source_failed'
        WHEN 'verification_failed' THEN 'verification_failed'
        WHEN 'sequence_conflict' THEN 'sequence_conflict'
        ELSE 'sequence_rollback'
    END;
    v_operation_state := CASE
        WHEN p_failure_kind IN ('source_unavailable', 'verification_failed')
            THEN 'pending'
        ELSE 'fenced'
    END;
    UPDATE platform.billing_license_refresh_attempts
    SET state = v_attempt_state, finished_at = v_now,
        failure_kind = p_failure_kind
    WHERE operation_id = p_operation_id
      AND attempt = v_operation.current_attempt
      AND lease_token = p_lease_token AND state = 'running';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'billing license refresh failure attempt is stale';
    END IF;
    UPDATE platform.billing_license_refresh_operations
    SET state = v_operation_state, lease_token = NULL, lease_until = NULL,
        updated_at = v_now,
        completed_at = CASE WHEN v_operation_state = 'fenced'
            THEN v_now ELSE NULL END
    WHERE operation_id = p_operation_id;
    RETURN v_operation_state;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_license_refresh_failure_v1(
    uuid, uuid, text
) FROM PUBLIC;


-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_platform_license_refresh_read_v1(
    p_environment_id uuid,
    p_operation_id uuid
)
RETURNS TABLE (
    operation_id uuid,
    operation_state text,
    snapshot_sequence bigint,
    snapshot_sha256 text,
    failure_code text,
    created_at timestamptz,
    completed_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.clock_timestamp();
    v_request platform.billing_platform_license_refresh_requests%ROWTYPE;
    v_operation platform.billing_license_refresh_operations%ROWTYPE;
    v_attempt_state text;
    v_attempt_failure text;
    v_attempt_finished timestamptz;
    v_receipt_sequence bigint;
    v_receipt_sha256 text;
    v_receipt_outcome text;
    v_state text;
    v_failure text;
    v_completed timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'harpy-v2:core-api' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing platform license refresh read denied';
    END IF;
    IF p_environment_id IS NULL
       OR pg_catalog.uuid_extract_version(p_environment_id) <> 7
       OR p_operation_id IS NULL
       OR pg_catalog.uuid_extract_version(p_operation_id) <> 7 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'billing platform license refresh read is invalid';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM platform.tenancy_deployment_policy AS policy
        WHERE policy.singleton
          AND policy.environment_id = p_environment_id
          AND policy.deployment_mode = 'onpremise'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'billing platform license refresh read requires onpremise';
    END IF;
    SELECT request.* INTO v_request
    FROM platform.billing_platform_license_refresh_requests AS request
    WHERE request.operation_id = p_operation_id
      AND request.environment_id = p_environment_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'billing platform license refresh operation was not found';
    END IF;
    SELECT refresh.* INTO v_operation
    FROM platform.billing_license_refresh_operations AS refresh
    WHERE refresh.operation_id = p_operation_id;
    IF NOT FOUND THEN
        RETURN QUERY SELECT p_operation_id, 'pending'::text,
            NULL::bigint, NULL::text, NULL::text,
            v_request.created_at, NULL::timestamptz;
        RETURN;
    END IF;
    SELECT attempt.state, attempt.failure_kind, attempt.finished_at
    INTO v_attempt_state, v_attempt_failure, v_attempt_finished
    FROM platform.billing_license_refresh_attempts AS attempt
    WHERE attempt.operation_id = p_operation_id
    ORDER BY attempt.attempt DESC
    LIMIT 1;
    SELECT receipt.sequence, receipt.payload_sha256, receipt.outcome
    INTO v_receipt_sequence, v_receipt_sha256, v_receipt_outcome
    FROM platform.billing_license_verification_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id
    ORDER BY receipt.attempt DESC
    LIMIT 1;
    IF v_operation.state = 'succeeded' THEN
        v_state := 'accepted';
        v_completed := v_operation.completed_at;
    ELSIF v_operation.state = 'fenced' THEN
        v_state := 'fenced';
        v_failure := COALESCE(v_attempt_failure, v_receipt_outcome);
        v_completed := v_operation.completed_at;
    ELSIF v_operation.state = 'running'
          AND v_operation.lease_until > v_now THEN
        v_state := 'leased';
    ELSIF v_operation.state = 'running' THEN
        v_state := 'failed';
        v_failure := 'outcome_unknown';
        v_completed := v_operation.lease_until;
    ELSIF v_operation.state = 'pending'
          AND v_attempt_failure IS NOT NULL THEN
        v_state := 'failed';
        v_failure := CASE
            WHEN v_attempt_failure = 'lease_expired' THEN 'outcome_unknown'
            ELSE v_attempt_failure
        END;
        v_completed := v_attempt_finished;
    ELSE
        v_state := 'pending';
    END IF;
    RETURN QUERY SELECT p_operation_id, v_state,
        v_receipt_sequence, v_receipt_sha256, v_failure,
        v_request.created_at, v_completed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.billing_platform_license_refresh_read_v1(uuid, uuid)
FROM PUBLIC;

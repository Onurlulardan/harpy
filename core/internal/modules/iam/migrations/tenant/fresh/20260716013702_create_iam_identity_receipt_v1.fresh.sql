-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_replay_v1(
    p_operation_id uuid, p_command_key text,
    p_request_sha256 bytea, p_result_id uuid
)
RETURNS TABLE (
    result_id uuid, result_revision bigint,
    authorization_clock bigint, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_command text; v_request bytea; v_id uuid;
    v_revision bigint; v_clock bigint;
BEGIN
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_id,
           receipt.result_revision, receipt.clock_value
    INTO v_command, v_request, v_id, v_revision, v_clock
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id;
    IF NOT FOUND THEN RETURN; END IF;
    IF v_command IS DISTINCT FROM p_command_key OR
       v_request IS DISTINCT FROM p_request_sha256 OR
       v_id IS DISTINCT FROM p_result_id THEN
        RAISE EXCEPTION USING ERRCODE = '23505',
            MESSAGE = 'IAM identity operation identity was reused';
    END IF;
    RETURN QUERY SELECT v_id, v_revision, v_clock, true;
END;
$function$;

CREATE FUNCTION iam_identity_record_v1(
    p_operation_id uuid, p_command_key text, p_request_sha256 bytea,
    p_result_id uuid, p_result_revision bigint, p_result_sha256 bytea
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_clock bigint;
BEGIN
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1, revision = revision + 1, updated_at = v_now
    WHERE singleton RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM identity authorization clock is unavailable';
    END IF;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value,
        result_code, result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, p_command_key, p_request_sha256, v_clock,
        p_command_key || '.succeeded', p_result_id, p_result_revision,
        p_result_sha256, v_now
    );
    RETURN v_clock;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_replay_v1(uuid,text,bytea,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_identity_record_v1(
    uuid,text,bytea,uuid,bigint,bytea
) FROM PUBLIC;

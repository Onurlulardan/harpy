-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_service_principal_create_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_principal_id uuid,
    p_principal_key text,
    p_display_name text,
    p_actor_id uuid,
    p_manage_permission_id uuid,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_purpose_sha256 bytea,
    p_auth_time timestamptz,
    p_amr_password boolean,
    p_amr_webauthn boolean,
    p_amr_totp boolean
)
RETURNS TABLE (
    result_id uuid,
    result_revision bigint,
    authorization_clock bigint,
    replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz;
    v_receipt_command text;
    v_receipt_request bytea;
    v_receipt_id uuid;
    v_receipt_revision bigint;
    v_receipt_clock bigint;
    v_clock bigint;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    IF p_operation_id IS NULL OR p_principal_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM service principal create input is invalid';
    END IF;
    IF NOT iam_service_principal_actor_assert_v1(
        p_scope_schema, p_actor_id, p_manage_permission_id,
        p_expected_token_version, p_expected_authorization_version,
        p_purpose_sha256, p_auth_time,
        p_amr_password, p_amr_webauthn, p_amr_totp
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM service principal command authority is denied';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_id,
           receipt.result_revision, receipt.clock_value
    INTO v_receipt_command, v_receipt_request, v_receipt_id,
         v_receipt_revision, v_receipt_clock
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id OPERATOR(pg_catalog.=) p_operation_id;
    IF FOUND THEN
        IF v_receipt_command IS DISTINCT FROM 'iam.service_principal.create' OR
           v_receipt_request IS DISTINCT FROM p_request_sha256 OR
           v_receipt_id IS DISTINCT FROM p_principal_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM service principal operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_receipt_id, v_receipt_revision,
            v_receipt_clock, true;
        RETURN;
    END IF;
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1,
        revision = revision + 1,
        updated_at = v_now
    WHERE singleton
    RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM authorization clock is unavailable';
    END IF;
    INSERT INTO iam_service_principals (
        id, principal_key, kind, display_name, state,
        authorization_version, revision, created_at, updated_at
    ) VALUES (
        p_principal_id, p_principal_key, 'workflow', p_display_name, 'active',
        v_clock, 1, v_now, v_now
    );
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value,
        result_code, result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, 'iam.service_principal.create', p_request_sha256,
        v_clock, 'iam.service_principal.created', p_principal_id, 1,
        p_request_sha256, v_now
    );
    RETURN QUERY SELECT p_principal_id, 1::bigint, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_service_principal_create_v1(
    name,uuid,bytea,uuid,text,text,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) FROM PUBLIC;

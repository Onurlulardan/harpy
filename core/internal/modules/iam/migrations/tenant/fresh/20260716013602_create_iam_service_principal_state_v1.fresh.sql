-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_service_principal_state_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_principal_id uuid,
    p_expected_revision bigint,
    p_target_state text,
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
    v_state text;
    v_revision bigint;
    v_command text;
    v_code text;
    v_receipt_command text;
    v_receipt_request bytea;
    v_receipt_id uuid;
    v_receipt_revision bigint;
    v_receipt_clock bigint;
    v_clock bigint;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    IF p_operation_id IS NULL OR p_principal_id IS NULL OR
       p_expected_revision < 1 OR
       p_target_state NOT IN ('active', 'disabled', 'retired') OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM service principal state input is invalid';
    END IF;
    v_command := CASE p_target_state
        WHEN 'active' THEN 'iam.service_principal.enable'
        WHEN 'disabled' THEN 'iam.service_principal.disable'
        ELSE 'iam.service_principal.retire'
    END;
    v_code := CASE p_target_state
        WHEN 'active' THEN 'iam.service_principal.enabled'
        WHEN 'disabled' THEN 'iam.service_principal.disabled'
        ELSE 'iam.service_principal.retired'
    END;
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
        IF v_receipt_command IS DISTINCT FROM v_command OR
           v_receipt_request IS DISTINCT FROM p_request_sha256 OR
           v_receipt_id IS DISTINCT FROM p_principal_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM service principal operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_receipt_id, v_receipt_revision,
            v_receipt_clock, true;
        RETURN;
    END IF;
    SELECT principal.state, principal.revision
    INTO v_state, v_revision
    FROM iam_service_principals AS principal
    WHERE principal.id OPERATOR(pg_catalog.=) p_principal_id
    FOR UPDATE;
    IF NOT FOUND OR v_revision IS DISTINCT FROM p_expected_revision OR
       NOT (
           (v_state = 'active' AND p_target_state IN ('disabled', 'retired')) OR
           (v_state = 'disabled' AND p_target_state IN ('active', 'retired'))
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM service principal state CAS is stale';
    END IF;
    IF p_target_state = 'retired' AND EXISTS (
        SELECT 1 FROM iam_capability_grants AS grant_head
        WHERE grant_head.principal_id = p_principal_id
          AND grant_head.state = 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23503',
            MESSAGE = 'IAM service principal has active capability grants';
    END IF;
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1,
        revision = revision + 1,
        updated_at = v_now
    WHERE singleton
    RETURNING clock_value INTO v_clock;
    UPDATE iam_service_principals
    SET state = p_target_state,
        authorization_version = v_clock,
        revision = v_revision + 1,
        updated_at = v_now,
        disabled_at = CASE
            WHEN p_target_state = 'active' THEN NULL
            WHEN p_target_state = 'disabled' THEN v_now
            ELSE disabled_at
        END,
        retired_at = CASE WHEN p_target_state = 'retired' THEN v_now ELSE NULL END
    WHERE id = p_principal_id;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value,
        result_code, result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, v_command, p_request_sha256, v_clock,
        v_code, p_principal_id, v_revision + 1, p_request_sha256, v_now
    );
    RETURN QUERY SELECT p_principal_id, v_revision + 1, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_service_principal_state_v1(
    name,uuid,bytea,uuid,bigint,text,uuid,uuid,bigint,bigint,
    bytea,timestamptz,boolean,boolean,boolean
) FROM PUBLIC;

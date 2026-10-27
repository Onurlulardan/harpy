-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_schema_publish_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_schema_id uuid,
    p_resource_kind text,
    p_version bigint,
    p_definition jsonb,
    p_definition_sha256 bytea,
    p_expected_active_schema_id uuid,
    p_actor_id uuid,
    p_manage_permission_id uuid,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint
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
    v_revision bigint;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    IF p_operation_id IS NULL OR p_schema_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile schema publish input is invalid';
    END IF;
    IF NOT iam_profile_actor_assert_v1(
        p_scope_schema, p_actor_id, p_manage_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile schema publish authority is denied';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_id,
           receipt.result_revision, receipt.clock_value
    INTO v_receipt_command, v_receipt_request, v_receipt_id,
         v_receipt_revision, v_receipt_clock
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id OPERATOR(pg_catalog.=) p_operation_id;
    IF FOUND THEN
        IF v_receipt_command IS DISTINCT FROM 'iam.profile_schema.publish' OR
           v_receipt_request IS DISTINCT FROM p_request_sha256 OR
           v_receipt_id IS DISTINCT FROM p_schema_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM profile schema publish identity was reused';
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
    v_revision := iam_profile_schema_apply_v1(
        p_scope_schema, p_schema_id, p_resource_kind, p_version,
        p_definition, p_definition_sha256, p_expected_active_schema_id,
        NULL, NULL
    );
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value,
        result_code, result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, 'iam.profile_schema.publish', p_request_sha256,
        v_clock, 'iam.profile_schema.published', p_schema_id, v_revision,
        p_definition_sha256, v_now
    );
    RETURN QUERY SELECT p_schema_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_schema_publish_v1(
    name,uuid,bytea,uuid,text,bigint,jsonb,bytea,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;


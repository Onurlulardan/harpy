-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_self_profile_update_v1(
    p_scope_schema name,
    p_operation_id uuid,
    p_request_sha256 bytea,
    p_actor_id uuid,
    p_expected_source_schema_id uuid,
    p_target_schema_id uuid,
    p_profile_document bytea,
    p_profile_sha256 bytea,
    p_expected_revision bigint,
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
    v_values jsonb;
    v_current_schema_id uuid;
    v_current_revision bigint;
    v_receipt_command text;
    v_receipt_request bytea;
    v_receipt_id uuid;
    v_receipt_revision bigint;
    v_receipt_clock bigint;
    v_clock bigint;
    v_revision bigint;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_operation_id IS NULL OR p_actor_id IS NULL OR
       p_expected_source_schema_id IS NULL OR p_target_schema_id IS NULL OR
       p_expected_revision < 1 OR p_expected_token_version < 1 OR
       p_expected_authorization_version < 1 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_profile_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_profile_document) > 1048576 OR
       pg_catalog.sha256(p_profile_document) IS DISTINCT FROM p_profile_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM self profile update input is invalid';
    END IF;
    SELECT actor.profile_schema_version_id, actor.revision
    INTO v_current_schema_id, v_current_revision
    FROM iam_users AS actor
    WHERE actor.id = p_actor_id
      AND actor.status = 'active'
      AND actor.token_version = p_expected_token_version
      AND actor.authorization_version = p_expected_authorization_version
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM self profile update authority is denied';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_id,
           receipt.result_revision, receipt.clock_value
    INTO v_receipt_command, v_receipt_request, v_receipt_id,
         v_receipt_revision, v_receipt_clock
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id;
    IF FOUND THEN
        IF v_receipt_command IS DISTINCT FROM 'iam.identity.user.update' OR
           v_receipt_request IS DISTINCT FROM p_request_sha256 OR
           v_receipt_id IS DISTINCT FROM p_actor_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM self profile operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_receipt_id, v_receipt_revision,
            v_receipt_clock, true;
        RETURN;
    END IF;
    IF v_current_schema_id IS DISTINCT FROM p_expected_source_schema_id OR
       v_current_revision IS DISTINCT FROM p_expected_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM self profile compare-and-swap failed';
    END IF;
    v_values := iam_identity_profile_assert_v1(
        'user', p_target_schema_id, p_profile_document, p_profile_sha256
    );
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1,
        revision = revision + 1,
        updated_at = v_now
    WHERE singleton
    RETURNING clock_value INTO v_clock;
    UPDATE iam_users
    SET profile_schema_version_id = p_target_schema_id,
        profile_values = v_values,
        profile_sha256 = p_profile_sha256,
        revision = revision + 1,
        updated_at = v_now
    WHERE id = p_actor_id AND revision = p_expected_revision
    RETURNING revision INTO v_revision;
    IF v_clock IS NULL OR v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM self profile update compare-and-swap failed';
    END IF;
    INSERT INTO iam_authorization_operation_receipts (
        operation_id, command_key, request_sha256, clock_value,
        result_code, result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, 'iam.identity.user.update', p_request_sha256,
        v_clock, 'iam.identity.user.update.succeeded', p_actor_id, v_revision,
        p_profile_sha256, v_now
    );
    RETURN QUERY SELECT p_actor_id, v_revision, v_clock, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_self_profile_update_v1(
    name,uuid,bytea,uuid,uuid,uuid,bytea,bytea,bigint,bigint,bigint
) FROM PUBLIC;

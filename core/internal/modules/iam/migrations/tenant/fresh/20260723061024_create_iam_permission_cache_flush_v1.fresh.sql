-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_cache_flush_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_actor_id uuid,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_expected_platform_clock bigint
)
RETURNS TABLE (
    authorization_clock bigint, result_revision bigint, replayed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_command text;
    v_request bytea;
    v_result_id uuid;
    v_clock bigint;
    v_revision bigint;
    v_status text;
    v_token_version bigint;
    v_authorization_version bigint;
    v_platform_allowed boolean;
BEGIN
    IF p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_tenant_id IS NULL OR p_operation_id IS NULL OR p_actor_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_expected_token_version < 1 OR p_expected_authorization_version < 1 OR
       p_expected_platform_clock < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM permission cache flush identity is invalid';
    END IF;

    SELECT user_record.status, user_record.token_version,
           user_record.authorization_version
    INTO v_status, v_token_version, v_authorization_version
    FROM iam_users AS user_record
    WHERE user_record.id = p_actor_id
    FOR UPDATE;
    IF NOT FOUND OR v_status IS DISTINCT FROM 'active' OR
       v_token_version IS DISTINCT FROM p_expected_token_version OR
       v_authorization_version IS DISTINCT FROM p_expected_authorization_version THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission cache flush actor is stale';
    END IF;

    SELECT decision.allowed INTO v_platform_allowed
    FROM platform.iam_platform_authorize_decide_v1(
        p_tenant_id, p_actor_id, p_expected_platform_clock,
        'platform.operations.permission_cache.flush'
    ) AS decision;
    IF v_platform_allowed IS DISTINCT FROM true THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM permission cache flush capability is denied';
    END IF;

    SELECT receipt.command_key, receipt.request_sha256, receipt.result_id,
           receipt.clock_value, receipt.result_revision
    INTO v_command, v_request, v_result_id, v_clock, v_revision
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id
    FOR UPDATE;
    IF FOUND THEN
        IF v_command IS DISTINCT FROM 'iam.permission_cache.flush' OR
           v_request IS DISTINCT FROM p_request_sha256 OR
           v_result_id IS DISTINCT FROM p_tenant_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM permission cache flush identity was reused';
        END IF;
        RETURN QUERY SELECT v_clock, v_revision, true;
        RETURN;
    END IF;

    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1, revision = revision + 1,
        updated_at = pg_catalog.transaction_timestamp()
    WHERE singleton
    RETURNING clock_value, revision INTO v_clock, v_revision;
    IF v_clock IS NULL OR v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM authorization clock is unavailable';
    END IF;

    INSERT INTO iam_authorization_operation_receipts(
        operation_id, command_key, request_sha256, clock_value,
        result_code, result_id, result_revision, result_sha256, created_at
    ) VALUES (
        p_operation_id, 'iam.permission_cache.flush', p_request_sha256, v_clock,
        'iam.permission_cache.flushed', p_tenant_id, v_revision,
        p_request_sha256, pg_catalog.transaction_timestamp()
    );
    RETURN QUERY SELECT v_clock, v_revision, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_cache_flush_v1(
    name,uuid,uuid,bytea,uuid,bigint,bigint,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_permission_cache_flush_v1(
    name,uuid,uuid,bytea,uuid,bigint,bigint,bigint
) TO forgecrud_runtime;


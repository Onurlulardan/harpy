-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_mutation_begin_v1(
    p_scope_schema name, p_operation_id uuid, p_command_key text,
    p_request_sha256 bytea, p_result_id uuid, p_actor_id uuid,
    p_required_permission_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint
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
    v_revision bigint;
    v_clock bigint;
    v_status text;
    v_token_version bigint;
    v_authorization_version bigint;
BEGIN
    IF p_operation_id IS NULL OR p_result_id IS NULL OR p_actor_id IS NULL OR
       p_required_permission_id IS NULL OR p_expected_token_version < 1 OR
       p_expected_authorization_version < 1 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       p_command_key !~ '^iam\.rbac\.[a-z0-9_.]+$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM RBAC mutation identity is invalid';
    END IF;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM RBAC mutation scope is invalid';
    END IF;
    SELECT status, token_version, authorization_version
    INTO v_status, v_token_version, v_authorization_version
    FROM iam_users WHERE id = p_actor_id FOR UPDATE;
    IF NOT FOUND OR v_status IS DISTINCT FROM 'active' OR
       NOT iam_user_has_permission_v1(p_actor_id, p_required_permission_id) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM RBAC mutation authority is denied';
    END IF;
    SELECT receipt.command_key, receipt.request_sha256, receipt.result_id,
        receipt.result_revision, receipt.clock_value
    INTO v_command, v_request, v_result_id, v_revision, v_clock
    FROM iam_authorization_operation_receipts AS receipt
    WHERE receipt.operation_id = p_operation_id FOR UPDATE;
    IF FOUND THEN
        IF v_command IS DISTINCT FROM p_command_key OR
           v_request IS DISTINCT FROM p_request_sha256 OR
           v_result_id IS DISTINCT FROM p_result_id THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM RBAC operation identity was reused';
        END IF;
        RETURN QUERY SELECT v_clock, v_revision, true;
        RETURN;
    END IF;
    IF v_token_version IS DISTINCT FROM p_expected_token_version OR
       v_authorization_version IS DISTINCT FROM p_expected_authorization_version THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM RBAC mutation actor version is stale';
    END IF;
    UPDATE iam_authorization_clock
    SET clock_value = clock_value + 1, revision = revision + 1,
        updated_at = pg_catalog.transaction_timestamp()
    WHERE singleton RETURNING clock_value INTO v_clock;
    IF v_clock IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM authorization clock is unavailable';
    END IF;
    RETURN QUERY SELECT v_clock, 0::bigint, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_mutation_begin_v1(
    name,uuid,text,bytea,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_mcp_authorization_request_create_v1(
    p_id uuid,
    p_operation_id text,
    p_request_sha256 bytea,
    p_client_id text,
    p_redirect_uri text,
    p_resource_uri text,
    p_scope text,
    p_state_value text,
    p_code_challenge text,
    p_code_challenge_method text
)
RETURNS TABLE (
    result_id uuid,
    result_status text,
    result_revision bigint,
    result_created_at timestamptz,
    result_expires_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_existing platform.iam_mcp_authorization_requests%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_id IS NULL OR p_operation_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP authorization request authority is invalid';
    END IF;
    INSERT INTO platform.iam_mcp_authorization_requests (
        id, operation_id, request_sha256, client_id, redirect_uri,
        resource_uri, scope, state_value, code_challenge,
        code_challenge_method, status, revision, created_at, expires_at,
        retain_until
    ) VALUES (
        p_id, p_operation_id, p_request_sha256, p_client_id, p_redirect_uri,
        p_resource_uri, p_scope, p_state_value, p_code_challenge,
        p_code_challenge_method, 'pending', 1, v_now,
        v_now + INTERVAL '10 minutes', v_now + INTERVAL '24 hours 10 minutes'
    )
    ON CONFLICT (operation_id) DO NOTHING;

    SELECT request.*
    INTO STRICT v_existing
    FROM platform.iam_mcp_authorization_requests AS request
    WHERE request.operation_id = p_operation_id
    FOR UPDATE;
    IF v_existing.request_sha256 IS DISTINCT FROM p_request_sha256 OR
       v_existing.client_id IS DISTINCT FROM p_client_id OR
       v_existing.redirect_uri IS DISTINCT FROM p_redirect_uri OR
       v_existing.resource_uri IS DISTINCT FROM p_resource_uri OR
       v_existing.scope IS DISTINCT FROM p_scope OR
       v_existing.state_value IS DISTINCT FROM p_state_value OR
       v_existing.code_challenge IS DISTINCT FROM p_code_challenge OR
       v_existing.code_challenge_method IS DISTINCT FROM
           p_code_challenge_method THEN
        RAISE EXCEPTION USING ERRCODE = 'P5711',
            MESSAGE = 'IAM MCP authorization request idempotency conflict';
    END IF;
    RETURN QUERY SELECT v_existing.id, v_existing.status,
        v_existing.revision, v_existing.created_at, v_existing.expires_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_mcp_authorization_request_create_v1(
    uuid,text,bytea,text,text,text,text,text,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_mcp_authorization_request_create_v1(
    uuid,text,bytea,text,text,text,text,text,text,text
) TO forgecrud_runtime;

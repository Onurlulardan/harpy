-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_mcp_code_exchange_snapshot_v1(
    p_scope_schema name,
    p_request_id uuid,
    p_grant_id uuid,
    p_code_selector bytea,
    p_code_secret_hmac bytea
)
RETURNS TABLE (
    result_user_id uuid,
    result_session_id uuid,
    result_client_id text,
    result_redirect_uri text,
    result_resource_uri text,
    result_scope text,
    result_code_challenge text,
    result_token_version bigint,
    result_authorization_version bigint,
    result_platform_clock bigint,
    result_security_epoch bytea,
    result_auth_time timestamptz,
    result_acr text,
    result_amr text[],
    result_issued_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.octet_length(p_code_selector) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_code_secret_hmac) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP code exchange authority is invalid';
    END IF;
    SELECT identity.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton AND identity.schema_name = p_scope_schema;

    RETURN QUERY
    SELECT oauth_grant.user_id, oauth_grant.session_id, oauth_grant.client_id,
           request.redirect_uri, oauth_grant.resource_uri, oauth_grant.scope,
           request.code_challenge, oauth_grant.token_version,
           oauth_grant.authorization_version, oauth_grant.platform_authorization_clock,
           oauth_grant.security_epoch, oauth_grant.auth_time, oauth_grant.acr, oauth_grant.amr,
           pg_catalog.transaction_timestamp()
    FROM platform.iam_mcp_authorization_requests AS request
    JOIN iam_mcp_oauth_grants AS oauth_grant
      ON oauth_grant.authorization_request_id = request.id
    JOIN iam_current_session_read_v6(
        p_scope_schema, oauth_grant.user_id, oauth_grant.session_id
    ) AS session
      ON session.result_token_version = oauth_grant.token_version
     AND session.result_authorization_version = oauth_grant.authorization_version
     AND session.result_platform_clock = oauth_grant.platform_authorization_clock
     AND session.result_security_epoch = oauth_grant.security_epoch
     AND session.result_client_id = 'harpy-operator-console-v1'
    WHERE request.id = p_request_id
      AND request.grant_id = p_grant_id
      AND request.tenant_id = v_tenant_id
      AND request.status = 'approved'
      AND request.code_selector = p_code_selector
      AND request.code_secret_hmac = p_code_secret_hmac
      AND request.code_expires_at > pg_catalog.transaction_timestamp()
      AND oauth_grant.id = p_grant_id
      AND oauth_grant.state = 'active'
      AND oauth_grant.absolute_expires_at > pg_catalog.transaction_timestamp();
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM MCP authorization code is invalid';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_mcp_code_exchange_snapshot_v1(
    name,uuid,uuid,bytea,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_mcp_code_exchange_snapshot_v1(
    name,uuid,uuid,bytea,bytea
) TO harpy_runtime;

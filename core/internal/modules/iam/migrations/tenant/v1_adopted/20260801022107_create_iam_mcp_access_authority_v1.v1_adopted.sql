-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_mcp_access_validate_v1(
    p_scope_schema name,
    p_grant_id uuid,
    p_access_jti uuid,
    p_access_token_sha256 bytea
)
RETURNS TABLE (
    result_user_id uuid,
    result_session_id uuid,
    result_client_id text,
    result_resource_uri text,
    result_token_version bigint,
    result_authorization_version bigint,
    result_platform_clock bigint,
    result_security_epoch bytea,
    result_auth_time timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.octet_length(p_access_token_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP access evidence authority is invalid';
    END IF;
    RETURN QUERY
    SELECT oauth_grant.user_id, oauth_grant.session_id, oauth_grant.client_id,
           oauth_grant.resource_uri, oauth_grant.token_version,
           oauth_grant.authorization_version, oauth_grant.platform_authorization_clock,
           oauth_grant.security_epoch, oauth_grant.auth_time
    FROM iam_mcp_oauth_access_tokens AS access
    JOIN iam_mcp_oauth_grants AS oauth_grant ON oauth_grant.id = access.grant_id
    JOIN iam_current_session_read_v6(
        p_scope_schema, oauth_grant.user_id, oauth_grant.session_id
    ) AS session
      ON session.result_token_version = oauth_grant.token_version
     AND session.result_authorization_version = oauth_grant.authorization_version
     AND session.result_platform_clock = oauth_grant.platform_authorization_clock
     AND session.result_security_epoch = oauth_grant.security_epoch
     AND session.result_client_id = 'forgecrud-operator-console-v1'
    WHERE access.jti = p_access_jti
      AND access.grant_id = p_grant_id
      AND access.token_sha256 = p_access_token_sha256
      AND access.state = 'active'
      AND access.expires_at > pg_catalog.transaction_timestamp()
      AND oauth_grant.id = p_grant_id
      AND oauth_grant.state = 'active'
      AND oauth_grant.absolute_expires_at > pg_catalog.transaction_timestamp();
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM MCP access evidence is invalid';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_mcp_access_validate_v1(
    name,uuid,uuid,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_mcp_access_validate_v1(
    name,uuid,uuid,bytea
) TO harpy_runtime;

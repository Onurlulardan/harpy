-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_mcp_refresh_snapshot_v1(
    p_scope_schema name,
    p_grant_id uuid,
    p_family_id uuid,
    p_token_id uuid,
    p_selector bytea,
    p_secret_hmac bytea
)
RETURNS TABLE (
    result_outcome text,
    result_source_revision bigint,
    result_sequence bigint,
    result_user_id uuid,
    result_session_id uuid,
    result_client_id text,
    result_resource_uri text,
    result_token_version bigint,
    result_authorization_version bigint,
    result_platform_clock bigint,
    result_security_epoch bytea,
    result_auth_time timestamptz,
    result_acr text,
    result_amr text[],
    result_issued_at timestamptz,
    result_successor_id uuid,
    result_access_jti uuid,
    result_access_token_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.octet_length(p_selector) <> 16 OR
       pg_catalog.octet_length(p_secret_hmac) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP refresh snapshot authority is invalid';
    END IF;

    RETURN QUERY
    SELECT CASE
               WHEN token.state = 'active' THEN 'rotate'
               WHEN token.state = 'consumed' AND
                    token.consumed_at + INTERVAL '30 seconds' >= v_now
                   THEN 'replay'
               WHEN token.state = 'consumed' THEN 'reuse'
           END,
           token.revision, token.sequence, oauth_grant.user_id, oauth_grant.session_id,
           oauth_grant.client_id, oauth_grant.resource_uri, oauth_grant.token_version,
           oauth_grant.authorization_version, oauth_grant.platform_authorization_clock,
           oauth_grant.security_epoch, oauth_grant.auth_time, oauth_grant.acr, oauth_grant.amr,
           CASE WHEN token.state = 'active' THEN v_now ELSE access.issued_at END,
           token.successor_id, access.jti, access.token_sha256
    FROM iam_mcp_oauth_refresh_tokens AS token
    JOIN iam_mcp_oauth_refresh_families AS family
      ON family.id = token.family_id
    JOIN iam_mcp_oauth_grants AS oauth_grant
      ON oauth_grant.id = family.grant_id
    JOIN iam_current_session_read_v6(
        p_scope_schema, oauth_grant.user_id, oauth_grant.session_id
    ) AS session
      ON session.result_token_version = oauth_grant.token_version
     AND session.result_authorization_version = oauth_grant.authorization_version
     AND session.result_platform_clock = oauth_grant.platform_authorization_clock
     AND session.result_security_epoch = oauth_grant.security_epoch
     AND session.result_client_id = 'forgecrud-operator-console-v1'
    LEFT JOIN iam_mcp_oauth_access_tokens AS access
      ON access.refresh_token_id = token.successor_id
    WHERE token.id = p_token_id AND token.family_id = p_family_id
      AND token.selector = p_selector AND token.secret_hmac = p_secret_hmac
      AND token.state IN ('active','consumed')
      AND token.idle_expires_at > v_now
      AND family.id = p_family_id AND family.grant_id = p_grant_id
      AND family.state = 'active' AND family.absolute_expires_at > v_now
      AND oauth_grant.id = p_grant_id AND oauth_grant.state = 'active'
      AND oauth_grant.absolute_expires_at > v_now;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM MCP refresh credential is invalid';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_mcp_refresh_snapshot_v1(
    name,uuid,uuid,uuid,bytea,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_mcp_refresh_snapshot_v1(
    name,uuid,uuid,uuid,bytea,bytea
) TO harpy_runtime;

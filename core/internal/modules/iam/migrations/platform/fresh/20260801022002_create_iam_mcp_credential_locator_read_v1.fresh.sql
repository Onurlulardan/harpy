-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_mcp_credential_locator_read_v1(
    p_selector bytea,
    p_credential_kind text
)
RETURNS TABLE (
    result_tenant_id uuid,
    result_user_id uuid,
    result_session_id uuid,
    result_request_id uuid,
    result_grant_id uuid,
    result_family_id uuid,
    result_token_id uuid,
    result_source_revision bigint,
    result_source_sha256 bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       p_credential_kind NOT IN ('authorization_code','refresh_token') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP credential locator authority is invalid';
    END IF;
    RETURN QUERY
    SELECT locator.tenant_id, request.user_id, request.session_id,
           locator.request_id, locator.grant_id,
           locator.family_id, locator.token_id, locator.source_revision,
           locator.source_sha256
    FROM platform.iam_mcp_credential_locators AS locator
    JOIN platform.iam_mcp_authorization_requests AS request
      ON request.grant_id = locator.grant_id
    WHERE locator.selector = p_selector
      AND locator.credential_kind = p_credential_kind
      AND locator.state = 'active'
      AND locator.retain_until > pg_catalog.transaction_timestamp();
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM MCP credential locator is unknown';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_mcp_credential_locator_read_v1(
    bytea,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_mcp_credential_locator_read_v1(
    bytea,text
) TO forgecrud_runtime;

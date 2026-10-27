-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_mcp_credential_target_resolve_v1(
    p_selector bytea,
    p_credential_kind text
)
RETURNS TABLE (
    result_tenant_id uuid,
    result_user_id uuid,
    result_request_id uuid,
    result_grant_id uuid,
    result_family_id uuid,
    result_token_id uuid,
    result_selector bytea,
    result_source_revision bigint,
    result_source_sha256 bytea,
    result_group_generation bigint,
    result_iam_generation bigint,
    result_layout_generation bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_tenant_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       p_credential_kind NOT IN ('authorization_code','refresh_token') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP credential target authority is invalid';
    END IF;

    SELECT locator.tenant_id
    INTO STRICT v_tenant_id
    FROM platform.iam_mcp_credential_locators AS locator
    WHERE locator.selector = p_selector
      AND locator.credential_kind = p_credential_kind
      AND ((p_credential_kind = 'authorization_code' AND locator.state = 'active') OR
           (p_credential_kind = 'refresh_token' AND
            locator.state IN ('active','consumed')))
      AND locator.retain_until > pg_catalog.transaction_timestamp();

    SELECT target.group_generation, target.iam_generation,
           target.layout_generation
    INTO STRICT result_group_generation, result_iam_generation,
         result_layout_generation
    FROM platform.iam_resolve_tenant_target_v1(v_tenant_id) AS target;

    SELECT locator.tenant_id, request.user_id, locator.request_id,
           locator.grant_id, locator.family_id, locator.token_id,
           locator.selector, locator.source_revision, locator.source_sha256
    INTO STRICT result_tenant_id, result_user_id, result_request_id,
         result_grant_id, result_family_id, result_token_id,
         result_selector, result_source_revision, result_source_sha256
    FROM platform.iam_mcp_credential_locators AS locator
    JOIN platform.iam_mcp_authorization_requests AS request
      ON request.grant_id = locator.grant_id
    WHERE locator.selector = p_selector
      AND locator.credential_kind = p_credential_kind
      AND locator.tenant_id = v_tenant_id
      AND ((p_credential_kind = 'authorization_code' AND locator.state = 'active') OR
           (p_credential_kind = 'refresh_token' AND
            locator.state IN ('active','consumed')))
      AND locator.retain_until > pg_catalog.transaction_timestamp()
    FOR UPDATE OF locator;

    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_mcp_credential_target_resolve_v1(
    bytea,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_mcp_credential_target_resolve_v1(
    bytea,text
) TO forgecrud_runtime;

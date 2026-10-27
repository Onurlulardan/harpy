-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_mcp_grant_revoke_v1(
    p_scope_schema name,
    p_grant_id uuid
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_revision bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_grant_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP grant revocation authority is invalid';
    END IF;
    SELECT oauth_grant.revision INTO STRICT v_revision
    FROM iam_mcp_oauth_grants AS oauth_grant
    WHERE oauth_grant.id = p_grant_id
    FOR UPDATE;
    IF EXISTS (
        SELECT 1 FROM iam_mcp_oauth_grants
        WHERE id = p_grant_id AND state = 'active'
    ) THEN
        UPDATE iam_mcp_oauth_access_tokens
        SET state = 'revoked', revoked_at = LEAST(v_now, expires_at),
            revision = revision + 1
        WHERE grant_id = p_grant_id AND state = 'active';
        UPDATE iam_mcp_oauth_refresh_tokens AS token
        SET state = 'revoked', revoked_at = v_now, updated_at = v_now,
            revision = token.revision + 1
        FROM iam_mcp_oauth_refresh_families AS family
        WHERE family.grant_id = p_grant_id AND token.family_id = family.id
          AND token.state = 'active';
        UPDATE iam_mcp_oauth_refresh_families
        SET state = 'revoked', revoked_at = v_now, updated_at = v_now,
            revision = revision + 1
        WHERE grant_id = p_grant_id AND state = 'active';
        UPDATE iam_mcp_oauth_grants
        SET state = 'revoked', revoked_at = v_now, updated_at = v_now,
            revision = revision + 1
        WHERE id = p_grant_id AND state = 'active'
        RETURNING revision INTO v_revision;
        UPDATE platform.iam_mcp_credential_locators AS locator
        SET state = 'revoked', source_revision = locator.source_revision + 1,
            source_sha256 = pg_catalog.sha256(
                locator.source_sha256 || pg_catalog.convert_to('revoked', 'UTF8')
            )
        WHERE locator.grant_id = p_grant_id AND locator.state = 'active';
        UPDATE platform.iam_mcp_authorization_requests
        SET status = 'revoked', revision = revision + 1
        WHERE grant_id = p_grant_id AND status = 'approved';
    END IF;
    RETURN v_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_mcp_grant_revoke_v1(name,uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_mcp_grant_revoke_v1(name,uuid)
    TO forgecrud_runtime;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_mcp_code_exchange_v1(
    p_scope_schema name,
    p_request_id uuid,
    p_grant_id uuid,
    p_code_selector bytea,
    p_code_secret_hmac bytea,
    p_access_jti uuid,
    p_access_token_sha256 bytea,
    p_refresh_family_id uuid,
    p_refresh_token_id uuid,
    p_refresh_selector bytea,
    p_refresh_hash_key_id text,
    p_refresh_secret_hmac bytea,
    p_code_locator_sha256 bytea,
    p_refresh_locator_sha256 bytea
)
RETURNS timestamptz
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_tenant_id uuid;
    v_request platform.iam_mcp_authorization_requests%ROWTYPE;
    v_grant iam_mcp_oauth_grants%ROWTYPE;
    v_token_version bigint;
    v_authorization_version bigint;
    v_platform_clock bigint;
    v_security_epoch bytea;
    v_client_id text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.octet_length(p_code_selector) <> 16 OR
       pg_catalog.octet_length(p_code_secret_hmac) <> 32 OR
       pg_catalog.octet_length(p_access_token_sha256) <> 32 OR
       pg_catalog.octet_length(p_refresh_selector) <> 16 OR
       pg_catalog.char_length(p_refresh_hash_key_id) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_refresh_secret_hmac) <> 32 OR
       pg_catalog.octet_length(p_code_locator_sha256) <> 32 OR
       pg_catalog.octet_length(p_refresh_locator_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP code exchange mutation is invalid';
    END IF;
    SELECT identity.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton AND identity.schema_name = p_scope_schema;

    SELECT request.* INTO STRICT v_request
    FROM platform.iam_mcp_authorization_requests AS request
    WHERE request.id = p_request_id
    FOR UPDATE;
    SELECT oauth_grant.* INTO STRICT v_grant
    FROM iam_mcp_oauth_grants AS oauth_grant
    WHERE oauth_grant.id = p_grant_id
    FOR UPDATE;

    IF v_request.status IS DISTINCT FROM 'approved' OR
       v_request.grant_id IS DISTINCT FROM p_grant_id OR
       v_request.tenant_id IS DISTINCT FROM v_tenant_id OR
       v_request.code_selector IS DISTINCT FROM p_code_selector OR
       v_request.code_secret_hmac IS DISTINCT FROM p_code_secret_hmac OR
       v_request.code_expires_at <= v_now OR
       v_grant.authorization_request_id IS DISTINCT FROM p_request_id OR
       v_grant.state IS DISTINCT FROM 'active' OR
       v_grant.absolute_expires_at <= v_now THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'IAM MCP authorization code is invalid';
    END IF;

    SELECT session.result_client_id, session.result_token_version,
           session.result_authorization_version,
           session.result_platform_clock, session.result_security_epoch
    INTO STRICT v_client_id, v_token_version, v_authorization_version,
         v_platform_clock, v_security_epoch
    FROM iam_current_session_read_v6(
        p_scope_schema, v_grant.user_id, v_grant.session_id
    ) AS session;
    IF v_client_id IS DISTINCT FROM 'forgecrud-operator-console-v1' OR
       v_token_version IS DISTINCT FROM v_grant.token_version OR
       v_authorization_version IS DISTINCT FROM v_grant.authorization_version OR
       v_platform_clock IS DISTINCT FROM v_grant.platform_authorization_clock OR
       v_security_epoch IS DISTINCT FROM v_grant.security_epoch THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP grant is stale';
    END IF;

    INSERT INTO iam_mcp_oauth_refresh_families (
        id, grant_id, state, rotation_sequence, revision, created_at,
        updated_at, absolute_expires_at, retain_until
    ) VALUES (
        p_refresh_family_id, p_grant_id, 'active', 1, 1, v_now,
        v_now, v_now + INTERVAL '90 days',
        v_now + INTERVAL '90 days 24 hours'
    );
    INSERT INTO iam_mcp_oauth_refresh_tokens (
        id, family_id, parent_id, sequence, selector, hash_key_id,
        secret_hmac, successor_id, state, revision, issued_at, updated_at,
        idle_expires_at, retain_until
    ) VALUES (
        p_refresh_token_id, p_refresh_family_id, NULL, 1,
        p_refresh_selector, p_refresh_hash_key_id, p_refresh_secret_hmac,
        NULL, 'active', 1, v_now, v_now, v_now + INTERVAL '30 days',
        v_now + INTERVAL '90 days 24 hours'
    );
    INSERT INTO iam_mcp_oauth_access_tokens (
        jti, grant_id, token_sha256, refresh_token_id, state, revision,
        issued_at, expires_at, retain_until
    ) VALUES (
        p_access_jti, p_grant_id, p_access_token_sha256,
        p_refresh_token_id, 'active', 1,
        v_now, v_now + INTERVAL '15 minutes',
        v_now + INTERVAL '24 hours 15 minutes'
    );

    UPDATE platform.iam_mcp_authorization_requests
    SET status = 'consumed', revision = revision + 1, consumed_at = v_now
    WHERE id = p_request_id;
    UPDATE platform.iam_mcp_credential_locators
    SET state = 'consumed', source_revision = source_revision + 1,
        source_sha256 = p_code_locator_sha256
    WHERE selector = p_code_selector
      AND credential_kind = 'authorization_code'
      AND state = 'active';

    INSERT INTO platform.iam_mcp_credential_locators (
        selector, credential_kind, tenant_id, grant_id, family_id, token_id,
        source_revision, source_sha256, state, created_at, retain_until
    ) VALUES (
        p_refresh_selector, 'refresh_token', v_tenant_id, p_grant_id,
        p_refresh_family_id, p_refresh_token_id, 1,
        p_refresh_locator_sha256, 'active', v_now,
        v_now + INTERVAL '90 days 24 hours'
    );
    RETURN v_now;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_mcp_code_exchange_v1(
    name,uuid,uuid,bytea,bytea,uuid,bytea,uuid,uuid,bytea,text,bytea,bytea,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_mcp_code_exchange_v1(
    name,uuid,uuid,bytea,bytea,uuid,bytea,uuid,uuid,bytea,text,bytea,bytea,bytea
) TO harpy_runtime;

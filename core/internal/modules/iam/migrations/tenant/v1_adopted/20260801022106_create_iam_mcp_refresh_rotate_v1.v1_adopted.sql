-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_mcp_refresh_rotate_v1(
    p_scope_schema name,
    p_grant_id uuid,
    p_family_id uuid,
    p_token_id uuid,
    p_expected_revision bigint,
    p_selector bytea,
    p_secret_hmac bytea,
    p_successor_id uuid,
    p_successor_selector bytea,
    p_successor_hash_key_id text,
    p_successor_secret_hmac bytea,
    p_access_jti uuid,
    p_access_token_sha256 bytea,
    p_consumed_locator_sha256 bytea,
    p_successor_locator_sha256 bytea
)
RETURNS timestamptz
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_token iam_mcp_oauth_refresh_tokens%ROWTYPE;
    v_family iam_mcp_oauth_refresh_families%ROWTYPE;
    v_grant iam_mcp_oauth_grants%ROWTYPE;
    v_idle_expires_at timestamptz;
    v_tenant_id uuid;
    v_existing_digest bytea;
    v_existing_issued_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       pg_catalog.octet_length(p_selector) <> 16 OR
       pg_catalog.octet_length(p_secret_hmac) <> 32 OR
       pg_catalog.octet_length(p_successor_selector) <> 16 OR
       pg_catalog.char_length(p_successor_hash_key_id) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_successor_secret_hmac) <> 32 OR
       pg_catalog.octet_length(p_access_token_sha256) <> 32 OR
       pg_catalog.octet_length(p_consumed_locator_sha256) <> 32 OR
       pg_catalog.octet_length(p_successor_locator_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP refresh rotation mutation is invalid';
    END IF;
    SELECT identity.tenant_id INTO STRICT v_tenant_id
    FROM tenancy_schema_identity AS identity
    WHERE identity.singleton AND identity.schema_name = p_scope_schema;
    SELECT token.* INTO STRICT v_token
    FROM iam_mcp_oauth_refresh_tokens AS token
    WHERE token.id = p_token_id
    FOR UPDATE;
    SELECT family.* INTO STRICT v_family
    FROM iam_mcp_oauth_refresh_families AS family
    WHERE family.id = p_family_id
    FOR UPDATE;
    SELECT oauth_grant.* INTO STRICT v_grant
    FROM iam_mcp_oauth_grants AS oauth_grant
    WHERE oauth_grant.id = p_grant_id
    FOR UPDATE;

    IF v_token.state = 'consumed' AND v_token.successor_id = p_successor_id AND
       v_token.consumed_at + INTERVAL '30 seconds' >= v_now THEN
        SELECT access.token_sha256, access.issued_at
        INTO STRICT v_existing_digest, v_existing_issued_at
        FROM iam_mcp_oauth_access_tokens AS access
        WHERE access.refresh_token_id = p_successor_id;
        IF v_existing_digest IS DISTINCT FROM p_access_token_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = 'P5711',
                MESSAGE = 'IAM MCP refresh replay response conflicts';
        END IF;
        RETURN v_existing_issued_at;
    END IF;
    IF v_token.state IS DISTINCT FROM 'active' OR
       v_token.revision IS DISTINCT FROM p_expected_revision OR
       v_token.family_id IS DISTINCT FROM p_family_id OR
       v_token.selector IS DISTINCT FROM p_selector OR
       v_token.secret_hmac IS DISTINCT FROM p_secret_hmac OR
       v_token.idle_expires_at <= v_now OR
       v_family.state IS DISTINCT FROM 'active' OR
       v_family.grant_id IS DISTINCT FROM p_grant_id OR
       v_family.absolute_expires_at <= v_now OR
       v_grant.state IS DISTINCT FROM 'active' OR
       v_grant.absolute_expires_at <= v_now THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM MCP refresh rotation is stale';
    END IF;
    PERFORM 1
    FROM iam_current_session_read_v6(
        p_scope_schema, v_grant.user_id, v_grant.session_id
    ) AS session
    WHERE session.result_token_version = v_grant.token_version
      AND session.result_authorization_version = v_grant.authorization_version
      AND session.result_platform_clock = v_grant.platform_authorization_clock
      AND session.result_security_epoch = v_grant.security_epoch
      AND session.result_client_id = 'forgecrud-operator-console-v1';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP refresh grant is stale';
    END IF;

    v_idle_expires_at := LEAST(
        v_now + INTERVAL '30 days', v_family.absolute_expires_at
    );
    UPDATE iam_mcp_oauth_refresh_tokens
    SET state = 'consumed', successor_id = p_successor_id,
        consumed_at = v_now, updated_at = v_now, revision = revision + 1
    WHERE id = p_token_id;
    INSERT INTO iam_mcp_oauth_refresh_tokens (
        id, family_id, parent_id, sequence, selector, hash_key_id,
        secret_hmac, successor_id, state, revision, issued_at, updated_at,
        idle_expires_at, retain_until
    ) VALUES (
        p_successor_id, p_family_id, p_token_id, v_token.sequence + 1,
        p_successor_selector, p_successor_hash_key_id,
        p_successor_secret_hmac, NULL, 'active', 1, v_now, v_now,
        v_idle_expires_at, v_family.retain_until
    );
    INSERT INTO iam_mcp_oauth_access_tokens (
        jti, grant_id, token_sha256, refresh_token_id, state, revision,
        issued_at, expires_at, retain_until
    ) VALUES (
        p_access_jti, p_grant_id, p_access_token_sha256, p_successor_id,
        'active', 1, v_now, v_now + INTERVAL '15 minutes',
        v_now + INTERVAL '24 hours 15 minutes'
    );
    UPDATE iam_mcp_oauth_refresh_families
    SET rotation_sequence = rotation_sequence + 1,
        updated_at = v_now, revision = revision + 1
    WHERE id = p_family_id;
    UPDATE platform.iam_mcp_credential_locators
    SET state = 'consumed', source_revision = source_revision + 1,
        source_sha256 = p_consumed_locator_sha256
    WHERE selector = p_selector AND credential_kind = 'refresh_token'
      AND state = 'active';
    INSERT INTO platform.iam_mcp_credential_locators (
        selector, credential_kind, tenant_id, grant_id, family_id, token_id,
        source_revision, source_sha256, state, created_at, retain_until
    ) VALUES (
        p_successor_selector, 'refresh_token', v_tenant_id, p_grant_id,
        p_family_id, p_successor_id, 1, p_successor_locator_sha256,
        'active', v_now, v_family.retain_until
    );
    RETURN v_now;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_mcp_refresh_rotate_v1(
    name,uuid,uuid,uuid,bigint,bytea,bytea,uuid,bytea,text,bytea,
    uuid,bytea,bytea,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_mcp_refresh_rotate_v1(
    name,uuid,uuid,uuid,bigint,bytea,bytea,uuid,bytea,text,bytea,
    uuid,bytea,bytea,bytea
) TO harpy_runtime;

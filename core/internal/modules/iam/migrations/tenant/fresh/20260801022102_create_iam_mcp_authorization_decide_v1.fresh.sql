-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_mcp_authorization_decide_v1(
    p_scope_schema name,
    p_request_id uuid,
    p_expected_revision bigint,
    p_decision_operation_id text,
    p_decision_sha256 bytea,
    p_approve boolean,
    p_user_id uuid,
    p_session_id uuid,
    p_auth_time timestamptz,
    p_acr text,
    p_amr text[],
    p_grant_id uuid,
    p_code_selector bytea,
    p_code_hash_key_id text,
    p_code_secret_hmac bytea,
    p_locator_sha256 bytea
)
RETURNS TABLE (
    result_status text,
    result_revision bigint,
    result_grant_id uuid
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_request platform.iam_mcp_authorization_requests%ROWTYPE;
    v_token_version bigint;
    v_authorization_version bigint;
    v_platform_clock bigint;
    v_security_epoch bytea;
    v_client_id text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_request_id IS NULL OR p_user_id IS NULL OR p_session_id IS NULL OR
       pg_catalog.octet_length(p_decision_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.char_length(p_decision_operation_id) NOT BETWEEN 1 AND 128 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP authorization decision authority is invalid';
    END IF;

    SELECT session.result_client_id, session.result_token_version,
           session.result_authorization_version,
           session.result_platform_clock, session.result_security_epoch
    INTO STRICT v_client_id, v_token_version, v_authorization_version,
         v_platform_clock, v_security_epoch
    FROM iam_current_session_read_v6(
        p_scope_schema, p_user_id, p_session_id
    ) AS session;
    IF v_client_id IS DISTINCT FROM 'forgecrud-operator-console-v1' OR
       p_auth_time IS NULL OR p_auth_time > v_now OR
       p_acr NOT IN ('urn:forgecrud:acr:baseline','urn:forgecrud:acr:step-up') OR
       pg_catalog.cardinality(p_amr) NOT BETWEEN 1 AND 3 OR
       NOT p_amr <@ ARRAY['pwd','webauthn','totp']::text[] THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM MCP authorization requires an operator session';
    END IF;

    SELECT request.*
    INTO STRICT v_request
    FROM platform.iam_mcp_authorization_requests AS request
    WHERE request.id = p_request_id
    FOR UPDATE;

    IF v_request.status IS DISTINCT FROM 'pending' THEN
        IF v_request.decision_operation_id IS NOT DISTINCT FROM
               p_decision_operation_id AND
           v_request.decision_sha256 IS NOT DISTINCT FROM p_decision_sha256 AND
           ((p_approve AND v_request.status IN ('approved','consumed','revoked')) OR
            (NOT p_approve AND v_request.status = 'denied')) THEN
            RETURN QUERY SELECT v_request.status, v_request.revision,
                v_request.grant_id;
            RETURN;
        END IF;
        RAISE EXCEPTION USING ERRCODE = 'P5711',
            MESSAGE = 'IAM MCP authorization decision conflicts';
    END IF;
    IF v_request.revision IS DISTINCT FROM p_expected_revision OR
       v_request.expires_at <= v_now THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM MCP authorization request is stale';
    END IF;

    IF NOT p_approve THEN
        UPDATE platform.iam_mcp_authorization_requests
        SET status = 'denied', revision = revision + 1,
            tenant_id = (SELECT tenant_id FROM tenancy_schema_identity
                         WHERE singleton AND schema_name = p_scope_schema),
            user_id = p_user_id, session_id = p_session_id,
            decision_operation_id = p_decision_operation_id,
            decision_sha256 = p_decision_sha256, decided_at = v_now
        WHERE id = p_request_id
        RETURNING status, revision, grant_id
        INTO result_status, result_revision, result_grant_id;
        RETURN NEXT;
        RETURN;
    END IF;

    IF p_grant_id IS NULL OR pg_catalog.octet_length(p_code_selector) <> 16 OR
       pg_catalog.char_length(p_code_hash_key_id) NOT BETWEEN 1 AND 128 OR
       pg_catalog.octet_length(p_code_secret_hmac) <> 32 OR
       pg_catalog.octet_length(p_locator_sha256) <> 32 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM MCP authorization code material is invalid';
    END IF;

    INSERT INTO iam_mcp_oauth_grants (
        id, authorization_request_id, user_id, session_id, client_id,
        resource_uri, scope, token_version, authorization_version,
        platform_authorization_clock, security_epoch, auth_time, acr, amr, state,
        revision, created_at, updated_at, absolute_expires_at, retain_until
    ) VALUES (
        p_grant_id, p_request_id, p_user_id, p_session_id,
        v_request.client_id, v_request.resource_uri, v_request.scope,
        v_token_version, v_authorization_version, v_platform_clock,
        v_security_epoch, p_auth_time, p_acr, p_amr, 'active', 1, v_now, v_now,
        v_now + INTERVAL '90 days', v_now + INTERVAL '90 days 24 hours'
    );

    UPDATE platform.iam_mcp_authorization_requests
    SET status = 'approved', revision = revision + 1,
        tenant_id = (SELECT tenant_id FROM tenancy_schema_identity
                     WHERE singleton AND schema_name = p_scope_schema),
        user_id = p_user_id, session_id = p_session_id, grant_id = p_grant_id,
        decision_operation_id = p_decision_operation_id,
        decision_sha256 = p_decision_sha256, code_selector = p_code_selector,
        code_hash_key_id = p_code_hash_key_id,
        code_secret_hmac = p_code_secret_hmac, decided_at = v_now,
        code_expires_at = v_now + INTERVAL '5 minutes'
    WHERE id = p_request_id;

    INSERT INTO platform.iam_mcp_credential_locators (
        selector, credential_kind, tenant_id, request_id, grant_id,
        source_revision, source_sha256, state, created_at, retain_until
    )
    SELECT p_code_selector, 'authorization_code', request.tenant_id,
           request.id, p_grant_id, request.revision, p_locator_sha256,
           'active', v_now, request.retain_until
    FROM platform.iam_mcp_authorization_requests AS request
    WHERE request.id = p_request_id;

    RETURN QUERY
    SELECT request.status, request.revision, request.grant_id
    FROM platform.iam_mcp_authorization_requests AS request
    WHERE request.id = p_request_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_mcp_authorization_decide_v1(
    name,uuid,bigint,text,bytea,boolean,uuid,uuid,timestamptz,text,text[],
    uuid,bytea,text,bytea,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_mcp_authorization_decide_v1(
    name,uuid,bigint,text,bytea,boolean,uuid,uuid,timestamptz,text,text[],
    uuid,bytea,text,bytea,bytea
) TO forgecrud_runtime;

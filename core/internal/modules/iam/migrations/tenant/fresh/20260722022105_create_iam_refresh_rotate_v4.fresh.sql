-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_rotate_v4(
    p_scope_schema name, p_token_id uuid, p_family_id uuid,
    p_expected_token_revision bigint, p_expected_family_revision bigint,
    p_session_id uuid, p_expected_session_revision bigint,
    p_user_id uuid, p_expected_user_revision bigint,
    p_expected_token_version bigint, p_expected_authorization_version bigint,
    p_access_mode_id uuid, p_expected_access_mode_revision bigint,
    p_portal_binding_id uuid, p_expected_portal_binding_revision bigint,
    p_successor_id uuid, p_successor_selector bytea,
    p_successor_hash_key_id text, p_successor_secret_hmac bytea,
    p_expected_platform_clock bigint, p_expected_security_epoch bytea
)
RETURNS TABLE (
    outcome text, issued_at timestamptz,
    locator_selector bytea, locator_expected_revision bigint,
    locator_revision bigint, locator_source_sha256 bytea,
    locator_state text, locator_projected_at timestamptz,
    locator_retain_until timestamptz, successor_selector bytea,
    successor_revision bigint, successor_source_sha256 bytea,
    successor_state text, successor_issued_at timestamptz,
    successor_projected_at timestamptz,
    successor_family_expires_at timestamptz,
    successor_retain_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_platform_clock bigint;
BEGIN
    IF p_expected_platform_clock < 0 OR
       pg_catalog.octet_length(p_expected_security_epoch) IS DISTINCT FROM 32 OR
       p_expected_security_epoch IS DISTINCT FROM
           iam_security_epoch_current_v1(p_scope_schema) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM refresh security evidence is stale';
    END IF;
    SELECT context.platform_clock INTO STRICT v_platform_clock
    FROM iam_platform_session_context_v1(
        p_scope_schema, p_user_id, p_expected_token_version,
        p_expected_authorization_version
    ) AS context;
    IF v_platform_clock IS DISTINCT FROM p_expected_platform_clock OR
       NOT EXISTS (
           SELECT 1 FROM iam_refresh_tokens AS token
           JOIN iam_refresh_families AS family ON family.id = token.family_id
           JOIN iam_sessions AS session ON session.id = family.session_id
           WHERE token.id = p_token_id AND family.id = p_family_id
             AND session.id = p_session_id
             AND token.security_epoch = p_expected_security_epoch
             AND family.security_epoch = p_expected_security_epoch
             AND session.security_epoch = p_expected_security_epoch
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM refresh security evidence is stale';
    END IF;
    RETURN QUERY SELECT * FROM iam_refresh_rotate_v3(
        p_scope_schema, p_token_id, p_family_id,
        p_expected_token_revision, p_expected_family_revision,
        p_session_id, p_expected_session_revision, p_user_id,
        p_expected_user_revision, p_expected_token_version,
        p_expected_authorization_version, p_access_mode_id,
        p_expected_access_mode_revision, p_portal_binding_id,
        p_expected_portal_binding_revision, p_successor_id,
        p_successor_selector, p_successor_hash_key_id, p_successor_secret_hmac
    );
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_refresh_rotate_v4(
    name,uuid,uuid,bigint,bigint,uuid,bigint,uuid,bigint,bigint,bigint,
    uuid,bigint,uuid,bigint,uuid,bytea,text,bytea,bigint,bytea
) FROM PUBLIC;

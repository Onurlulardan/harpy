-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_session_issue_v2(
    p_scope_schema name, p_user_id uuid, p_expected_user_revision bigint,
    p_expected_token_version bigint, p_expected_authorization_version bigint,
    p_access_mode_id uuid, p_access_mode text,
    p_expected_access_mode_revision bigint, p_portal_binding_id uuid,
    p_expected_portal_binding_revision bigint, p_session_id uuid,
    p_device_kind text, p_browser_family text, p_os_family text,
    p_ip_hmac_key_id text, p_ip_hmac bytea, p_family_id uuid,
    p_token_id uuid, p_selector bytea, p_hash_key_id text, p_secret_hmac bytea
)
RETURNS TABLE (
    session_revision bigint, family_revision bigint, token_revision bigint,
    issued_at timestamptz, idle_expires_at timestamptz,
    absolute_expires_at timestamptz, retain_until timestamptz,
    platform_authorization_clock bigint, security_epoch bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    SELECT issued.session_revision, issued.family_revision,
           issued.token_revision, issued.issued_at,
           issued.idle_expires_at, issued.absolute_expires_at,
           issued.retain_until
    INTO STRICT session_revision, family_revision, token_revision,
         issued_at, idle_expires_at, absolute_expires_at, retain_until
    FROM iam_session_issue_v1(
        p_scope_schema, p_user_id, p_expected_user_revision,
        p_expected_token_version, p_expected_authorization_version,
        p_access_mode_id, p_access_mode, p_expected_access_mode_revision,
        p_portal_binding_id, p_expected_portal_binding_revision, p_session_id,
        p_device_kind, p_browser_family, p_os_family, p_ip_hmac_key_id,
        p_ip_hmac, p_family_id, p_token_id, p_selector, p_hash_key_id,
        p_secret_hmac
    ) AS issued;
    SELECT context.platform_clock INTO STRICT platform_authorization_clock
    FROM iam_platform_session_context_v1(
        p_scope_schema, p_user_id, p_expected_token_version,
        p_expected_authorization_version
    ) AS context;
    SELECT session.security_epoch INTO STRICT security_epoch
    FROM iam_sessions AS session
    JOIN iam_refresh_families AS family ON family.session_id = session.id
    JOIN iam_refresh_tokens AS token ON token.family_id = family.id
    WHERE session.id = p_session_id AND family.id = p_family_id
      AND token.id = p_token_id
      AND session.security_epoch = family.security_epoch
      AND family.security_epoch = token.security_epoch
      AND token.security_epoch = iam_security_epoch_current_v1(p_scope_schema);
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_session_issue_v2(
    name,uuid,bigint,bigint,bigint,uuid,text,bigint,uuid,bigint,
    uuid,text,text,text,text,bytea,uuid,uuid,bytea,text,bytea
) FROM PUBLIC;

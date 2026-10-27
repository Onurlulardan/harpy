-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_credential_v3(
    p_scope_schema name, p_source_token_id uuid, p_source_family_id uuid,
    p_selector bytea, p_source_revision bigint, p_source_sha256 bytea
)
RETURNS TABLE (
    token_id uuid, token_revision bigint, token_state text,
    family_id uuid, family_revision bigint, family_state text,
    session_id uuid, session_revision bigint, session_state text,
    user_id uuid, user_revision bigint, user_status text, identity_kind text,
    token_version bigint, authorization_version bigint,
    access_mode_id uuid, access_mode text, access_mode_state text,
    access_mode_revision bigint, portal_binding_id uuid,
    portal_binding_state text, portal_binding_revision bigint,
    hash_key_id text, secret_hmac bytea, sequence bigint,
    auth_time timestamptz, acr text, amr_password boolean,
    amr_webauthn boolean, amr_totp boolean, amr_recovery boolean,
    step_up_purpose_sha256 bytea, idle_expires_at timestamptz,
    absolute_expires_at timestamptz, retain_until timestamptz,
    platform_authorization_clock bigint, security_epoch bytea
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN QUERY
    SELECT credential.*, context.platform_clock, token.security_epoch
    FROM iam_refresh_credential_v2(
        p_scope_schema, p_source_token_id, p_source_family_id,
        p_selector, p_source_revision, p_source_sha256
    ) AS credential
    JOIN iam_refresh_tokens AS token ON token.id = credential.token_id
    JOIN iam_refresh_families AS family ON family.id = credential.family_id
    JOIN iam_sessions AS session ON session.id = credential.session_id
    CROSS JOIN iam_platform_session_context_v1(
        p_scope_schema, credential.user_id, credential.token_version,
        credential.authorization_version
    ) AS context
    WHERE token.security_epoch = family.security_epoch
      AND family.security_epoch = session.security_epoch
      AND session.security_epoch = iam_security_epoch_current_v1(p_scope_schema);
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_refresh_credential_v3(
    name,uuid,uuid,bytea,bigint,bytea
) FROM PUBLIC;

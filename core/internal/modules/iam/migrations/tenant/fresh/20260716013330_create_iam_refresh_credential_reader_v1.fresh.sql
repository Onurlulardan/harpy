-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_refresh_credential_v1(
    p_scope_schema name,
    p_source_token_id uuid,
    p_source_family_id uuid,
    p_selector bytea,
    p_source_revision bigint,
    p_source_sha256 bytea
)
RETURNS TABLE (
    token_id uuid,
    token_revision bigint,
    token_state text,
    family_id uuid,
    family_revision bigint,
    family_state text,
    session_id uuid,
    session_revision bigint,
    session_state text,
    user_id uuid,
    user_revision bigint,
    user_status text,
    identity_kind text,
    token_version bigint,
    authorization_version bigint,
    access_mode_id uuid,
    access_mode text,
    access_mode_state text,
    access_mode_revision bigint,
    portal_binding_id uuid,
    portal_binding_state text,
    portal_binding_revision bigint,
    hash_key_id text,
    secret_hmac bytea,
    sequence bigint,
    auth_time timestamptz,
    acr text,
    amr_password boolean,
    amr_webauthn boolean,
    amr_totp boolean,
    amr_recovery boolean,
    step_up_purpose_sha256 bytea,
    idle_expires_at timestamptz,
    absolute_expires_at timestamptz,
    retain_until timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_source_token_id IS NULL OR p_source_family_id IS NULL OR
       pg_catalog.octet_length(p_selector) IS DISTINCT FROM 16 OR
       p_source_revision OPERATOR(pg_catalog.<=) 0 OR
       pg_catalog.octet_length(p_source_sha256) IS DISTINCT FROM 32 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM refresh credential scope is invalid';
    END IF;
    RETURN QUERY
    SELECT token.id, token.revision, token.state,
           family.id, family.revision, family.state,
           session.id, session.revision, session.state,
           principal.id, principal.revision, principal.status,
           principal.identity_kind, principal.token_version,
           principal.authorization_version, access.id, session.access_mode,
           access.state, access.revision, session.portal_binding_id,
           binding.state, binding.revision, token.hash_key_id, token.secret_hmac,
           token.sequence, token.auth_time, token.acr,
           token.amr_password, token.amr_webauthn, token.amr_totp,
           token.amr_recovery, token.step_up_purpose_sha256,
           token.idle_expires_at, family.absolute_expires_at,
           token.retain_until
    FROM iam_refresh_tokens AS token
    JOIN iam_refresh_families AS family
      ON family.id OPERATOR(pg_catalog.=) token.family_id
    JOIN iam_sessions AS session
      ON session.id OPERATOR(pg_catalog.=) family.session_id
    JOIN iam_users AS principal
      ON principal.id OPERATOR(pg_catalog.=) family.user_id
    JOIN iam_user_access_modes AS access
      ON access.id OPERATOR(pg_catalog.=) session.access_mode_id
     AND access.user_id OPERATOR(pg_catalog.=) principal.id
     AND access.mode OPERATOR(pg_catalog.=) session.access_mode
    LEFT JOIN iam_portal_subject_bindings AS binding
      ON binding.id OPERATOR(pg_catalog.=) session.portal_binding_id
     AND binding.user_id OPERATOR(pg_catalog.=) principal.id
     AND binding.access_mode_id OPERATOR(pg_catalog.=) session.access_mode_id
    WHERE token.id OPERATOR(pg_catalog.=) p_source_token_id
      AND token.family_id OPERATOR(pg_catalog.=) p_source_family_id
      AND family.id OPERATOR(pg_catalog.=) p_source_family_id
      AND token.selector OPERATOR(pg_catalog.=) p_selector
      AND token.revision OPERATOR(pg_catalog.=) p_source_revision
      AND token.secret_hmac OPERATOR(pg_catalog.=) p_source_sha256
      AND family.user_id OPERATOR(pg_catalog.=) principal.id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_refresh_credential_v1(
    name,uuid,uuid,bytea,bigint,bytea
) FROM PUBLIC;

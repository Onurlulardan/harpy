-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_reauthentication_start_v1(
    p_scope_schema name, p_tenant_id uuid, p_scope_binding bytea,
    p_challenge_id uuid, p_user_id uuid, p_session_id uuid,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_access_mode text, p_portal_binding_id uuid, p_method text,
    p_authenticator_id uuid, p_hmac_key_id text, p_challenge_hmac bytea,
    p_context_sha256 bytea, p_purpose_sha256 bytea,
    p_action_policy_version bigint, p_origin_sha256 bytea,
    p_require_phishing_resistant boolean
)
RETURNS TABLE (
    revision bigint, issued_at timestamptz, expires_at timestamptz,
    credential_id bytea, rp_id text, user_verification_required boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_session_id uuid;
    v_authenticator_kind text;
    v_origin_policy_version bigint;
    v_rp_id_sha256 bytea;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_challenge_id IS NULL OR p_user_id IS NULL OR p_session_id IS NULL OR
       pg_catalog.octet_length(p_scope_binding) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_challenge_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_context_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_purpose_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_origin_sha256) IS DISTINCT FROM 32 OR
       p_method NOT IN ('password_v1', 'webauthn_v1', 'totp_v1') OR
       (p_require_phishing_resistant AND
        p_method IS DISTINCT FROM 'webauthn_v1') OR
       iam_reauth_binding_sha256_v1(
           p_tenant_id, p_user_id, p_session_id, p_expected_token_version,
           p_expected_authorization_version, p_access_mode,
           p_portal_binding_id
       ) IS DISTINCT FROM p_scope_binding OR
       iam_reauth_context_sha256_v1(
           p_tenant_id, p_user_id, p_session_id, p_purpose_sha256,
           p_action_policy_version, p_method, p_origin_sha256
       ) IS DISTINCT FROM p_context_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM reauthentication start evidence is invalid';
    END IF;
    SELECT session.id INTO v_session_id
    FROM iam_sessions AS session
    JOIN iam_users AS principal
      ON principal.id OPERATOR(pg_catalog.=) session.user_id
    JOIN iam_user_access_modes AS access
      ON access.id OPERATOR(pg_catalog.=) session.access_mode_id
     AND access.user_id OPERATOR(pg_catalog.=) principal.id
     AND access.mode OPERATOR(pg_catalog.=) session.access_mode
    JOIN iam_refresh_families AS family
      ON family.session_id OPERATOR(pg_catalog.=) session.id
    JOIN iam_refresh_tokens AS token
      ON token.family_id OPERATOR(pg_catalog.=) family.id
     AND token.state OPERATOR(pg_catalog.=) 'active'
    LEFT JOIN iam_portal_subject_bindings AS binding
      ON binding.id OPERATOR(pg_catalog.=) session.portal_binding_id
     AND binding.user_id OPERATOR(pg_catalog.=) principal.id
     AND binding.access_mode_id OPERATOR(pg_catalog.=) access.id
    WHERE session.id OPERATOR(pg_catalog.=) p_session_id
      AND principal.id OPERATOR(pg_catalog.=) p_user_id
      AND principal.status OPERATOR(pg_catalog.=) 'active'
      AND principal.identity_kind OPERATOR(pg_catalog.=) 'human'
      AND principal.token_version OPERATOR(pg_catalog.=)
          p_expected_token_version
      AND principal.authorization_version OPERATOR(pg_catalog.=)
          p_expected_authorization_version
      AND session.state OPERATOR(pg_catalog.=) 'active'
      AND session.access_mode OPERATOR(pg_catalog.=) p_access_mode
      AND session.portal_binding_id IS NOT DISTINCT FROM p_portal_binding_id
      AND access.state OPERATOR(pg_catalog.=) 'active'
      AND family.state OPERATOR(pg_catalog.=) 'active'
      AND session.absolute_expires_at OPERATOR(pg_catalog.>) v_now
      AND family.absolute_expires_at OPERATOR(pg_catalog.>) v_now
      AND token.idle_expires_at OPERATOR(pg_catalog.>) v_now
      AND ((p_access_mode OPERATOR(pg_catalog.=) 'operator_console' AND
            binding.id IS NULL) OR
           (p_access_mode OPERATOR(pg_catalog.=) 'portal' AND
            binding.state OPERATOR(pg_catalog.=) 'active'));
    IF v_session_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM reauthentication start context is stale';
    END IF;
    IF p_method OPERATOR(pg_catalog.=) 'password_v1' THEN
        IF p_authenticator_id IS NOT NULL OR NOT EXISTS (
            SELECT 1 FROM iam_password_credentials AS password
            WHERE password.user_id OPERATOR(pg_catalog.=) p_user_id
              AND password.state OPERATOR(pg_catalog.=) 'active'
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM password reauthentication is unavailable';
        END IF;
    ELSIF p_method OPERATOR(pg_catalog.=) 'webauthn_v1' THEN
        SELECT authenticator.kind, webauthn.origin_policy_version,
               webauthn.rp_id_sha256, webauthn.credential_id,
               webauthn.rp_id, webauthn.uv_required
        INTO v_authenticator_kind, v_origin_policy_version,
             v_rp_id_sha256, credential_id, rp_id,
             user_verification_required
        FROM iam_authenticators AS authenticator
        JOIN iam_webauthn_credentials AS webauthn
          ON webauthn.authenticator_id OPERATOR(pg_catalog.=) authenticator.id
        WHERE authenticator.id OPERATOR(pg_catalog.=) p_authenticator_id
          AND authenticator.user_id OPERATOR(pg_catalog.=) p_user_id
          AND authenticator.kind OPERATOR(pg_catalog.=) 'webauthn_v1'
          AND authenticator.state OPERATOR(pg_catalog.=) 'active'
          AND webauthn.rp_id_sha256 OPERATOR(pg_catalog.=)
              pg_catalog.sha256(pg_catalog.convert_to(webauthn.rp_id, 'UTF8'));
        IF v_authenticator_kind IS NULL THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM WebAuthn reauthentication is unavailable';
        END IF;
    ELSE
        SELECT authenticator.kind INTO v_authenticator_kind
        FROM iam_authenticators AS authenticator
        JOIN iam_totp_authenticators AS totp
          ON totp.authenticator_id OPERATOR(pg_catalog.=) authenticator.id
        WHERE authenticator.id OPERATOR(pg_catalog.=) p_authenticator_id
          AND authenticator.user_id OPERATOR(pg_catalog.=) p_user_id
          AND authenticator.kind OPERATOR(pg_catalog.=) 'totp_v1'
          AND authenticator.state OPERATOR(pg_catalog.=) 'active'
          AND totp.profile OPERATOR(pg_catalog.=) 'totp_sha256_6_30_v1'
          AND totp.verified_at IS NOT NULL;
        IF v_authenticator_kind IS NULL THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM TOTP reauthentication is unavailable';
        END IF;
    END IF;
    issued_at := v_now;
    expires_at := v_now OPERATOR(pg_catalog.+) INTERVAL '5 minutes';
    revision := 1;
    INSERT INTO iam_auth_challenges (
        id, user_id, session_id, authenticator_id, authenticator_kind,
        purpose, proof_kind, hash_profile, hmac_key_id, challenge_hmac,
        context_sha256, step_up_purpose_sha256, origin_policy_version,
        action_policy_version, rp_id_sha256, origin_sha256,
        user_verification_required, state, attempt_count, revision,
        issued_at, expires_at, terminal_at, purge_after, created_at, updated_at
    ) VALUES (
        p_challenge_id, p_user_id, p_session_id, p_authenticator_id,
        v_authenticator_kind, 'reauth_v1', p_method, 'hmac_sha256_v1',
        p_hmac_key_id, p_challenge_hmac, p_context_sha256,
        p_purpose_sha256, v_origin_policy_version, p_action_policy_version,
        v_rp_id_sha256, CASE WHEN p_method = 'webauthn_v1'
            THEN p_origin_sha256 END, user_verification_required,
        'pending', 0, 1, v_now, expires_at, NULL,
        expires_at OPERATOR(pg_catalog.+) INTERVAL '24 hours', v_now, v_now
    );
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reauthentication_start_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,bigint,text,uuid,text,uuid,text,
    bytea,bytea,bytea,bigint,bytea,boolean
) FROM PUBLIC;

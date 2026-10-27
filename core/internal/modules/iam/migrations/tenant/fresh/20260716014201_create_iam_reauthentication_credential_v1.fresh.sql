-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_reauthentication_credential_v1(
    p_scope_schema name, p_tenant_id uuid, p_scope_binding bytea,
    p_challenge_id uuid, p_user_id uuid, p_session_id uuid,
    p_expected_token_version bigint,
    p_expected_authorization_version bigint,
    p_access_mode text, p_portal_binding_id uuid, p_method text,
    p_purpose_sha256 bytea, p_action_policy_version bigint,
    p_origin_sha256 bytea, p_context_sha256 bytea
)
RETURNS TABLE (
    challenge_revision bigint, attempt_count smallint, challenge_state text,
    hmac_key_id text, challenge_hmac bytea, context_sha256 bytea,
    purpose_sha256 bytea, origin_sha256 bytea, action_policy_version bigint,
    expires_at timestamptz, user_revision bigint, token_version bigint,
    authorization_version bigint, user_status text, identity_kind text,
    session_revision bigint, session_state text,
    session_absolute_expires_at timestamptz, access_mode_id uuid,
    access_mode text, access_mode_state text, access_mode_revision bigint,
    portal_binding_id uuid, portal_binding_state text,
    portal_binding_revision bigint, family_id uuid, family_revision bigint,
    family_state text, family_absolute_expires_at timestamptz,
    token_id uuid, token_revision bigint, token_state text,
    password_credential_version bigint, password_credential_revision bigint,
    password_algorithm text, password_salt bytea, password_tag bytea,
    password_memory_kib integer, password_iterations smallint,
    password_parallelism smallint, password_legacy_hash text,
    password_created_at timestamptz, authenticator_id uuid,
    authenticator_state text, authenticator_kind text,
    authenticator_revision bigint, webauthn_credential_id bytea,
    webauthn_public_key_cose bytea, webauthn_rp_id text,
    webauthn_rp_id_sha256 bytea, webauthn_origin_policy_version bigint,
    webauthn_sign_count bigint, webauthn_backup_eligible boolean,
    webauthn_backup_state boolean, webauthn_uv_required boolean,
    webauthn_revision bigint, totp_profile text, totp_cipher_suite text,
    totp_kms_key_id text, totp_nonce bytea, totp_secret_ciphertext bytea,
    totp_authentication_tag bytea, totp_aad_sha256 bytea,
    totp_last_accepted_time_step bigint, totp_revision bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_rows bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_challenge_id IS NULL OR p_user_id IS NULL OR p_session_id IS NULL OR
       pg_catalog.octet_length(p_scope_binding) IS DISTINCT FROM 32 OR
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
            MESSAGE = 'IAM reauthentication credential evidence is invalid';
    END IF;
    RETURN QUERY
    SELECT challenge.revision, challenge.attempt_count, challenge.state,
           challenge.hmac_key_id, challenge.challenge_hmac,
           challenge.context_sha256, challenge.step_up_purpose_sha256,
           p_origin_sha256, challenge.action_policy_version,
           challenge.expires_at, principal.revision,
           principal.token_version, principal.authorization_version,
           principal.status, principal.identity_kind, session.revision,
           session.state, session.absolute_expires_at, access.id,
           session.access_mode, access.state, access.revision, binding.id,
           binding.state, binding.revision, family.id, family.revision,
           family.state, family.absolute_expires_at, token.id,
           token.revision, token.state, password.credential_version,
           password.revision, password.algorithm, password.salt,
           password.tag, password.memory_kib, password.iterations,
           password.parallelism, password.legacy_bcrypt_hash,
           password.created_at, authenticator.id, authenticator.state,
           authenticator.kind, authenticator.revision,
           webauthn.credential_id, webauthn.public_key_cose, webauthn.rp_id,
           webauthn.rp_id_sha256, webauthn.origin_policy_version,
           webauthn.sign_count, webauthn.backup_eligible,
           webauthn.backup_state, webauthn.uv_required, webauthn.revision,
           totp.profile, totp.cipher_suite, totp.kms_key_id, totp.nonce,
           totp.secret_ciphertext, totp.authentication_tag, totp.aad_sha256,
           totp.last_accepted_time_step, totp.revision
    FROM iam_auth_challenges AS challenge
    JOIN iam_users AS principal
      ON principal.id OPERATOR(pg_catalog.=) challenge.user_id
    JOIN iam_sessions AS session
      ON session.id OPERATOR(pg_catalog.=) challenge.session_id
     AND session.user_id OPERATOR(pg_catalog.=) principal.id
    JOIN iam_user_access_modes AS access
      ON access.id OPERATOR(pg_catalog.=) session.access_mode_id
     AND access.user_id OPERATOR(pg_catalog.=) principal.id
     AND access.mode OPERATOR(pg_catalog.=) session.access_mode
    JOIN iam_refresh_families AS family
      ON family.session_id OPERATOR(pg_catalog.=) session.id
     AND family.user_id OPERATOR(pg_catalog.=) principal.id
    JOIN iam_refresh_tokens AS token
      ON token.family_id OPERATOR(pg_catalog.=) family.id
     AND token.state OPERATOR(pg_catalog.=) 'active'
    LEFT JOIN iam_portal_subject_bindings AS binding
      ON binding.id OPERATOR(pg_catalog.=) session.portal_binding_id
     AND binding.user_id OPERATOR(pg_catalog.=) principal.id
     AND binding.access_mode_id OPERATOR(pg_catalog.=) access.id
    LEFT JOIN iam_password_credentials AS password
      ON password.user_id OPERATOR(pg_catalog.=) principal.id
     AND challenge.proof_kind OPERATOR(pg_catalog.=) 'password_v1'
    LEFT JOIN iam_authenticators AS authenticator
      ON authenticator.id OPERATOR(pg_catalog.=) challenge.authenticator_id
     AND authenticator.user_id OPERATOR(pg_catalog.=) principal.id
     AND authenticator.kind OPERATOR(pg_catalog.=) challenge.authenticator_kind
    LEFT JOIN iam_webauthn_credentials AS webauthn
      ON webauthn.authenticator_id OPERATOR(pg_catalog.=) authenticator.id
     AND challenge.proof_kind OPERATOR(pg_catalog.=) 'webauthn_v1'
    LEFT JOIN iam_totp_authenticators AS totp
      ON totp.authenticator_id OPERATOR(pg_catalog.=) authenticator.id
     AND challenge.proof_kind OPERATOR(pg_catalog.=) 'totp_v1'
    WHERE challenge.id OPERATOR(pg_catalog.=) p_challenge_id
      AND challenge.user_id OPERATOR(pg_catalog.=) p_user_id
      AND challenge.session_id OPERATOR(pg_catalog.=) p_session_id
      AND challenge.purpose OPERATOR(pg_catalog.=) 'reauth_v1'
      AND challenge.proof_kind OPERATOR(pg_catalog.=) p_method
      AND challenge.context_sha256 OPERATOR(pg_catalog.=) p_context_sha256
      AND challenge.step_up_purpose_sha256 OPERATOR(pg_catalog.=)
          p_purpose_sha256
      AND challenge.action_policy_version OPERATOR(pg_catalog.=)
          p_action_policy_version
      AND principal.token_version OPERATOR(pg_catalog.=)
          p_expected_token_version
      AND principal.authorization_version OPERATOR(pg_catalog.=)
          p_expected_authorization_version
      AND session.access_mode OPERATOR(pg_catalog.=) p_access_mode
      AND session.portal_binding_id IS NOT DISTINCT FROM p_portal_binding_id
      AND ((p_method OPERATOR(pg_catalog.=) 'webauthn_v1' AND
            challenge.origin_sha256 OPERATOR(pg_catalog.=) p_origin_sha256) OR
           (p_method OPERATOR(pg_catalog.<>) 'webauthn_v1' AND
            challenge.origin_sha256 IS NULL));
    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows OPERATOR(pg_catalog.>) 1 THEN
        RAISE EXCEPTION USING ERRCODE = '21000',
            MESSAGE = 'IAM reauthentication credential is ambiguous';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_reauthentication_credential_v1(
    name,uuid,bytea,uuid,uuid,uuid,bigint,bigint,text,uuid,text,bytea,
    bigint,bytea,bytea
) FROM PUBLIC;

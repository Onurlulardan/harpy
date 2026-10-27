-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_totp_enrollment_start_v1(
    p_scope_schema name, p_tenant_id uuid, p_operation_id uuid,
    p_request_sha256 bytea, p_authenticator_id uuid, p_challenge_id uuid,
    p_user_id uuid, p_session_id uuid, p_display_label text,
    p_kms_key_id text, p_nonce bytea, p_secret_ciphertext bytea,
    p_authentication_tag bytea, p_aad_sha256 bytea,
    p_hmac_key_id text, p_challenge_hmac bytea, p_context_sha256 bytea,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    result_id uuid, subject_user_id uuid, result_revision bigint,
    token_version bigint, authorization_version bigint, replayed boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_token_version bigint;
    v_authorization_version bigint;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF p_authenticator_id IS NULL OR p_challenge_id IS NULL OR
       p_user_id IS NULL OR p_session_id IS NULL OR
       pg_catalog.octet_length(p_nonce) IS DISTINCT FROM 12 OR
       pg_catalog.octet_length(p_secret_ciphertext) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_authentication_tag) IS DISTINCT FROM 16 OR
       pg_catalog.octet_length(p_aad_sha256) IS DISTINCT FROM 32 OR
       p_hmac_key_id !~ '^[a-z0-9][a-z0-9._-]{0,127}$' OR
       pg_catalog.octet_length(p_challenge_hmac) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_context_sha256) IS DISTINCT FROM 32 OR
       NOT EXISTS (SELECT 1 FROM platform.tenants AS tenant
          WHERE tenant.id = p_tenant_id AND tenant.schema_name = p_scope_schema::text) THEN
        RAISE EXCEPTION USING ERRCODE = '22023';
    END IF;
    RETURN QUERY SELECT * FROM iam_lifecycle_replay_v1(
        p_operation_id, 'iam.totp_enrollment.start',
        p_request_sha256, p_authenticator_id
    );
    IF FOUND THEN RETURN; END IF;
    IF p_user_id <> p_actor_id OR NOT iam_identity_actor_assert_v1(
        p_scope_schema, p_actor_id, p_permission_id,
        p_expected_token_version, p_expected_authorization_version
    ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    IF NOT EXISTS (
        SELECT 1 FROM iam_sessions AS session
        WHERE session.id = p_session_id AND session.user_id = p_user_id
          AND session.state = 'active' AND session.absolute_expires_at > v_now
        FOR UPDATE
    ) THEN RAISE EXCEPTION USING ERRCODE = '28000'; END IF;
    INSERT INTO iam_authenticators (
        id, user_id, kind, display_label, state, revision, created_at, updated_at
    ) VALUES (
        p_authenticator_id, p_user_id, 'totp_v1', p_display_label,
        'pending', 1, v_now, v_now
    );
    INSERT INTO iam_totp_authenticators (
        authenticator_id, user_id, authenticator_kind, profile, cipher_suite,
        kms_key_id, nonce, secret_ciphertext, authentication_tag, aad_sha256,
        enrollment_expires_at, revision, created_at, updated_at
    ) VALUES (
        p_authenticator_id, p_user_id, 'totp_v1', 'totp_sha256_6_30_v1',
        'aes_256_gcm_v1', p_kms_key_id, p_nonce, p_secret_ciphertext,
        p_authentication_tag, p_aad_sha256, v_now + INTERVAL '10 minutes',
        1, v_now, v_now
    );
    INSERT INTO iam_auth_challenges (
        id, user_id, session_id, authenticator_id, authenticator_kind,
        purpose, proof_kind, hash_profile, hmac_key_id, challenge_hmac,
        context_sha256, state, attempt_count, revision, issued_at, expires_at,
        purge_after, created_at, updated_at
    ) VALUES (
        p_challenge_id, p_user_id, p_session_id, p_authenticator_id, 'totp_v1',
        'totp_enroll_v1', 'totp_v1', 'hmac_sha256_v1', p_hmac_key_id,
        p_challenge_hmac, p_context_sha256, 'pending', 0, 1, v_now,
        v_now + INTERVAL '5 minutes', v_now + INTERVAL '1 day 5 minutes',
        v_now, v_now
    );
    SELECT result.token_version, result.authorization_version
    INTO STRICT v_token_version, v_authorization_version
    FROM iam_lifecycle_record_v1(
        p_operation_id, 'iam.totp_enrollment.start', p_request_sha256,
        p_authenticator_id, p_user_id, 1, 'actor_tenant', p_actor_id,
        'authenticator', false, false
    ) AS result;
    RETURN QUERY SELECT p_authenticator_id, p_user_id, 1::bigint,
        v_token_version, v_authorization_version, false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_totp_enrollment_start_v1(
    name,uuid,uuid,bytea,uuid,uuid,uuid,uuid,text,text,bytea,bytea,bytea,bytea,
    text,bytea,bytea,uuid,uuid,bigint,bigint
) FROM PUBLIC;

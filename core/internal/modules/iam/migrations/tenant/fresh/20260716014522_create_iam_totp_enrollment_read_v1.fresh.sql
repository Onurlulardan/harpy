-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_totp_enrollment_read_v1(
    p_scope_schema name, p_authenticator_id uuid, p_challenge_id uuid,
    p_actor_id uuid, p_permission_id uuid,
    p_expected_token_version bigint, p_expected_authorization_version bigint
)
RETURNS TABLE (
    user_id uuid, kms_key_id text, nonce bytea, secret_ciphertext bytea,
    authentication_tag bytea, aad_sha256 bytea, authenticator_revision bigint,
    challenge_hmac_key_id text, challenge_hmac bytea, context_sha256 bytea,
    challenge_revision bigint, enrollment_expires_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF p_authenticator_id IS NULL OR p_challenge_id IS NULL OR
       NOT iam_identity_actor_assert_v1(
          p_scope_schema, p_actor_id, p_permission_id,
          p_expected_token_version, p_expected_authorization_version
       ) THEN RAISE EXCEPTION USING ERRCODE = '42501'; END IF;
    RETURN QUERY
    SELECT authenticator.user_id, totp.kms_key_id, totp.nonce,
        totp.secret_ciphertext, totp.authentication_tag, totp.aad_sha256,
        authenticator.revision, challenge.hmac_key_id,
        challenge.challenge_hmac, challenge.context_sha256,
        challenge.revision, totp.enrollment_expires_at
    FROM iam_authenticators AS authenticator
    JOIN iam_totp_authenticators AS totp
      ON totp.authenticator_id = authenticator.id
     AND totp.user_id = authenticator.user_id
    JOIN iam_auth_challenges AS challenge
      ON challenge.id = p_challenge_id
     AND challenge.authenticator_id = authenticator.id
     AND challenge.user_id = authenticator.user_id
    WHERE authenticator.id = p_authenticator_id
      AND authenticator.user_id = p_actor_id
      AND authenticator.kind = 'totp_v1'
      AND authenticator.state = 'pending'
      AND challenge.purpose = 'totp_enroll_v1'
      AND challenge.state = 'pending'
      AND challenge.expires_at > pg_catalog.transaction_timestamp()
      AND totp.enrollment_expires_at > pg_catalog.transaction_timestamp();
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_totp_enrollment_read_v1(
    name,uuid,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;


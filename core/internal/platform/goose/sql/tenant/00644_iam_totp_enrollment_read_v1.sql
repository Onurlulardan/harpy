-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_totp_enrollment_read_v1(
    in_schema_name name,
    in_authenticator_id uuid,
    in_challenge_id uuid,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    user_id uuid,
    kms_key_id text,
    nonce bytea,
    ciphertext bytea,
    authentication_tag bytea,
    aad bytea,
    authenticator_revision bigint,
    challenge_hmac_key_id text,
    challenge_hmac bytea,
    challenge_context bytea,
    challenge_revision bigint,
    expires_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::bytea,
        ''::bytea,
        ''::bytea,
        ''::bytea,
        0::bigint,
        ''::text,
        ''::bytea,
        ''::bytea,
        0::bigint,
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_totp_enrollment_read_v1(name, uuid, uuid, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_totp_enrollment_read_v1(name, uuid, uuid, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_totp_enrollment_read_v1(name, uuid, uuid, uuid, uuid, bigint, bigint);

-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_totp_enrollment_start_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_authenticator_id uuid,
    in_id uuid,
    in_user_id uuid,
    in_session_id uuid,
    in_display_label text,
    in_kms_key_id text,
    in_nonce bytea,
    in_ciphertext bytea,
    in_authentication_tag bytea,
    in_aadsha256 bytea,
    in_hmac_key_id text,
    in_hmac bytea,
    in_context_sha256 bytea,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint
)
RETURNS TABLE (
    id uuid,
    subject_user_id uuid,
    revision bigint,
    token_version bigint,
    authorization_version bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint,
        false
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_totp_enrollment_start_v1(name, uuid, uuid, bytea, uuid, uuid, uuid, uuid, text, text, bytea, bytea, bytea, bytea, text, bytea, bytea, uuid, uuid, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_totp_enrollment_start_v1(name, uuid, uuid, bytea, uuid, uuid, uuid, uuid, text, text, bytea, bytea, bytea, bytea, text, bytea, bytea, uuid, uuid, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_totp_enrollment_start_v1(name, uuid, uuid, bytea, uuid, uuid, uuid, uuid, text, text, bytea, bytea, bytea, bytea, text, bytea, bytea, uuid, uuid, bigint, bigint);

-- +goose Up
CREATE FUNCTION iam_email_verification_start_v2(
    in_schema_name name,
    in_tenant_id uuid,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_id bytea,
    in_identifier_id uuid,
    in_selector bytea,
    in_hash_key_id text,
    in_secret_hmac bytea,
    in_binding_sha256 bytea,
    in_actor_id uuid,
    in_permission_id uuid,
    in_expected_token_version bigint,
    in_expected_authorization_version bigint,
    in_artifact_id uuid,
    in_subject_user_id uuid,
    in_resource_id uuid,
    in_recipient_identifier_id uuid,
    in_reference text,
    in_key_id text,
    in_nonce bytea,
    in_ciphertext bytea,
    in_tag bytea,
    in_aadsha256 bytea,
    in_ciphertext_sha256 bytea
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
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_subject_user_id,
        0::bigint,
        0::bigint,
        0::bigint,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_email_verification_start_v2(name, uuid, uuid, bytea, bytea, uuid, bytea, text, bytea, bytea, uuid, uuid, bigint, bigint, uuid, uuid, uuid, uuid, text, text, bytea, bytea, bytea, bytea, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_email_verification_start_v2(name, uuid, uuid, bytea, bytea, uuid, bytea, text, bytea, bytea, uuid, uuid, bigint, bigint, uuid, uuid, uuid, uuid, text, text, bytea, bytea, bytea, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_email_verification_start_v2(name, uuid, uuid, bytea, bytea, uuid, bytea, text, bytea, bytea, uuid, uuid, bigint, bigint, uuid, uuid, uuid, uuid, text, text, bytea, bytea, bytea, bytea, bytea);

-- +goose Up
CREATE FUNCTION iam_invitation_create_v2(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_id uuid,
    in_display_value text,
    in_normalized_value text,
    in_normalized_sha256 bytea,
    in_role_id uuid,
    in_kind text,
    in_id_2 uuid,
    in_organization_id uuid,
    in_invited_by uuid,
    in_expected_authorization_version bigint,
    in_id_3 uuid,
    in_selector bytea,
    in_hash_key_id text,
    in_secret_hmac bytea,
    in_binding_sha256 bytea,
    in_artifact_id uuid,
    in_delivery_artifact_ref text,
    in_key_id text,
    in_nonce bytea,
    in_ciphertext bytea,
    in_tag bytea,
    in_aadsha256 bytea,
    in_ciphertext_sha256 bytea
)
RETURNS TABLE (
    id uuid,
    state text,
    revision bigint,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_id,
        ''::text,
        0::bigint,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_invitation_create_v2(name, uuid, bytea, uuid, text, text, bytea, uuid, text, uuid, uuid, uuid, bigint, uuid, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_create_v2(name, uuid, bytea, uuid, text, text, bytea, uuid, text, uuid, uuid, uuid, bigint, uuid, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_create_v2(name, uuid, bytea, uuid, text, text, bytea, uuid, text, uuid, uuid, uuid, bigint, uuid, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea);

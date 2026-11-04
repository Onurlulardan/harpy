-- +goose Up
CREATE FUNCTION platform.iam_initial_admin_setup_stage_v1(
    in_operation_id uuid,
    in_tenant_id uuid,
    in_user_id uuid,
    in_identifier_id uuid,
    in_email text,
    in_id bytea,
    in_selector bytea,
    in_hash_key_id text,
    in_secret_hmac bytea,
    in_binding_sha256 bytea,
    in_artifact_id uuid,
    in_reference text,
    in_key_id text,
    in_nonce bytea,
    in_ciphertext bytea,
    in_tag bytea,
    in_aadsha256 bytea,
    in_ciphertext_sha256 bytea,
    in_created_at timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.iam_initial_admin_setup_stage_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_initial_admin_setup_stage_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_initial_admin_setup_stage_v1(uuid, uuid, uuid, uuid, text, bytea, bytea, text, bytea, bytea, uuid, text, text, bytea, bytea, bytea, bytea, bytea, timestamptz);

-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_system_import_source_register_v2(
    in_schema_name name,
    in_source_id uuid,
    in_receipt_id uuid,
    in_actor_id uuid,
    in_resource text,
    in_token_version bigint,
    in_authorization_version bigint,
    in_idempotency_key_sha256 bytea,
    in_format text,
    in_object_key text,
    in_object_version text,
    in_header_sha256 bytea,
    in_ciphertext_sha256 bytea,
    in_ciphertext_bytes bigint,
    in_plaintext_sha256 bytea,
    in_plaintext_bytes bigint,
    in_key_id text,
    in_wrapped_dek bytea,
    in_wrap_nonce bytea,
    in_provider_request_id text,
    in_uploaded_at timestamptz
)
RETURNS TABLE (
    source_id uuid,
    resource_kind text,
    format text,
    plaintext_bytes bigint,
    plaintext_sha256 bytea,
    uploaded_at timestamptz,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_source_id,
        ''::text,
        in_format,
        in_plaintext_bytes,
        in_plaintext_sha256,
        in_uploaded_at,
        false
    ;
$$;

REVOKE ALL ON FUNCTION iam_system_import_source_register_v2(name, uuid, uuid, uuid, text, bigint, bigint, bytea, text, text, text, bytea, bytea, bigint, bytea, bigint, text, bytea, bytea, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_system_import_source_register_v2(name, uuid, uuid, uuid, text, bigint, bigint, bytea, text, text, text, bytea, bytea, bigint, bytea, bigint, text, bytea, bytea, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_system_import_source_register_v2(name, uuid, uuid, uuid, text, bigint, bigint, bytea, text, text, text, bytea, bytea, bigint, bytea, bigint, text, bytea, bytea, text, timestamptz);

-- +goose Up
CREATE FUNCTION platform.platformops_record_recovery_component_v3(
    in_recovery_set_id text,
    in_attempt_id text,
    in_lease_token text,
    in_kind text,
    in_object_key text,
    in_object_version text,
    in_cipher_sha256 text,
    in_ciphertext_bytes bigint
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformops_record_recovery_component_v3(text, text, text, text, text, text, text, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformops_record_recovery_component_v3(text, text, text, text, text, text, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformops_record_recovery_component_v3(text, text, text, text, text, text, text, bigint);

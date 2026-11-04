-- +goose Up
CREATE FUNCTION platform.platformauthority_prepare_v2(
    in_operation_id text,
    in_registry_binding_id text,
    in_purpose text,
    in_scope_kind text,
    in_scope_id text,
    in_generation bigint,
    in_previous_head_sha256 text,
    in_sha256 text,
    in_bytes bytea,
    in_sha256_10 text,
    in_bytes_11 bytea,
    in_compensates_ack_sha256 text,
    in_issued_at timestamptz,
    in_expires_at timestamptz,
    in_signer_kid text,
    in_now timestamptz
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.platformauthority_prepare_v2(text, text, text, text, text, bigint, text, text, bytea, text, bytea, text, timestamptz, timestamptz, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_prepare_v2(text, text, text, text, text, bigint, text, text, bytea, text, bytea, text, timestamptz, timestamptz, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_prepare_v2(text, text, text, text, text, bigint, text, text, bytea, text, bytea, text, timestamptz, timestamptz, text, timestamptz);

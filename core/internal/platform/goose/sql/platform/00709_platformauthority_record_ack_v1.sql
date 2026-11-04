-- +goose Up
CREATE FUNCTION platform.platformauthority_record_ack_v1(
    in_operation_id text,
    in_registry_binding_id text,
    in_scope_kind text,
    in_scope_id text,
    in_generation bigint,
    in_previous_head_sha256 text,
    in_state_sha256 text,
    in_proposal_sha256 text,
    in_sha256 text,
    in_bytes bytea,
    in_registry_sequence bigint,
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

REVOKE ALL ON FUNCTION platform.platformauthority_record_ack_v1(text, text, text, text, bigint, text, text, text, text, bytea, bigint, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platformauthority_record_ack_v1(text, text, text, text, bigint, text, text, text, text, bytea, bigint, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_record_ack_v1(text, text, text, text, bigint, text, text, text, text, bytea, bigint, timestamptz);

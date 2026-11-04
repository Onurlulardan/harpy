-- +goose Up
CREATE FUNCTION iam_identity_change_audit_write_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_actor_id uuid,
    in_resource_kind text,
    in_aggregate_id uuid,
    in_request_sha256 bytea,
    in_command_key text,
    in_result_code text
)
RETURNS TABLE (
    aggregate_revision bigint,
    authorization_clock bigint,
    event_sequence bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        0::bigint,
        0::bigint,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_identity_change_audit_write_v1(name, uuid, uuid, text, uuid, bytea, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identity_change_audit_write_v1(name, uuid, uuid, text, uuid, bytea, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identity_change_audit_write_v1(name, uuid, uuid, text, uuid, bytea, text, text);

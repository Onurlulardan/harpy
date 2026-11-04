-- +goose Up
CREATE FUNCTION iam_permission_request_transition_audit_write_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_group_id uuid,
    in_actor_id uuid,
    in_review_task uuid,
    in_command text,
    in_from_state text,
    in_to text,
    in_revision bigint
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION iam_permission_request_transition_audit_write_v1(name, uuid, bytea, uuid, uuid, uuid, text, text, text, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_permission_request_transition_audit_write_v1(name, uuid, bytea, uuid, uuid, uuid, text, text, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_permission_request_transition_audit_write_v1(name, uuid, bytea, uuid, uuid, uuid, text, text, text, bigint);

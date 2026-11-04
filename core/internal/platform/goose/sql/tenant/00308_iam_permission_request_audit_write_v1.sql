-- +goose Up
CREATE FUNCTION iam_permission_request_audit_write_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_request_sha256 bytea,
    in_group_id uuid,
    in_reviewer_id uuid,
    in_review_task_id uuid,
    in_decision text,
    in_code text,
    in_revision bigint,
    in_authorization_clock bigint
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION iam_permission_request_audit_write_v1(name, uuid, bytea, uuid, uuid, uuid, text, text, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_permission_request_audit_write_v1(name, uuid, bytea, uuid, uuid, uuid, text, text, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_permission_request_audit_write_v1(name, uuid, bytea, uuid, uuid, uuid, text, text, bigint, bigint);

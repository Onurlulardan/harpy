-- +goose Up
CREATE FUNCTION iam_invitation_audit_write_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_actor_id uuid,
    in_invitation_id uuid,
    in_request_sha256 bytea,
    in_command_key text,
    in_result_state text,
    in_result_revision bigint,
    in_artifact_ref text,
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

REVOKE ALL ON FUNCTION iam_invitation_audit_write_v1(name, uuid, uuid, uuid, bytea, text, text, bigint, text, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_invitation_audit_write_v1(name, uuid, uuid, uuid, bytea, text, text, bigint, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_invitation_audit_write_v1(name, uuid, uuid, uuid, bytea, text, text, bigint, text, bigint);

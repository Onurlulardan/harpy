-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_audit_write_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_actor_id uuid,
    in_aggregate_id uuid,
    in_request_sha256 bytea,
    in_command_key text,
    in_result_code text,
    in_revision bigint,
    in_clock bigint
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_audit_write_v1(name, uuid, uuid, uuid, bytea, text, text, bigint, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rbac_audit_write_v1(name, uuid, uuid, uuid, bytea, text, text, bigint, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rbac_audit_write_v1(name, uuid, uuid, uuid, bytea, text, text, bigint, bigint);

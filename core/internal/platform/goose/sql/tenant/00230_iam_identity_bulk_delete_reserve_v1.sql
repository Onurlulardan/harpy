-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_bulk_delete_reserve_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_actor_id uuid,
    in_resource_kind text,
    in_request_sha256 bytea,
    in_deleted_count integer
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

REVOKE ALL ON FUNCTION iam_identity_bulk_delete_reserve_v1(name, uuid, uuid, text, bytea, integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_identity_bulk_delete_reserve_v1(name, uuid, uuid, text, bytea, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_identity_bulk_delete_reserve_v1(name, uuid, uuid, text, bytea, integer);

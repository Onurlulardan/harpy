-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_delete_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 bigint
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

REVOKE ALL ON FUNCTION system_resource_layout_delete_v1(name, uuid, bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_delete_v1(name, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_delete_v1(name, uuid, bigint);

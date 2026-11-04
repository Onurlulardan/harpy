-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_list_v1(
    arg_1 name,
    arg_2 text
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::bytea;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_list_v1(name, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_list_v1(name, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_list_v1(name, text);

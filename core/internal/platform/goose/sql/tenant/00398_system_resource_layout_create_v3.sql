-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_create_v3(
    arg_1 name,
    arg_2 uuid,
    arg_3 text,
    arg_4 text,
    arg_5 bytea,
    arg_6 timestamptz
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

REVOKE ALL ON FUNCTION system_resource_layout_create_v3(name, uuid, text, text, bytea, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_create_v3(name, uuid, text, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_create_v3(name, uuid, text, text, bytea, timestamptz);

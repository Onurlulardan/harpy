-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_publish_v5(
    arg_1 name,
    arg_2 uuid,
    arg_3 bigint,
    arg_4 bytea,
    arg_5 text,
    arg_6 text,
    arg_7 text,
    arg_8 bytea,
    arg_9 text,
    arg_10 text,
    arg_11 bytea,
    arg_12 timestamptz
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

REVOKE ALL ON FUNCTION system_resource_layout_publish_v5(name, uuid, bigint, bytea, text, text, text, bytea, text, text, bytea, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_publish_v5(name, uuid, bigint, bytea, text, text, text, bytea, text, text, bytea, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION system_resource_layout_publish_v5(name, uuid, bigint, bytea, text, text, text, bytea, text, text, bytea, timestamptz);

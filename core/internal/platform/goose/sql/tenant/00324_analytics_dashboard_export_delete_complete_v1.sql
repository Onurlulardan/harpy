-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_delete_complete_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 bigint,
    arg_6 text,
    arg_7 text,
    arg_8 bytea,
    arg_9 bytea,
    arg_10 timestamptz,
    arg_11 timestamptz
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

REVOKE ALL ON FUNCTION analytics_dashboard_export_delete_complete_v1(name, uuid, uuid, uuid, bigint, text, text, bytea, bytea, timestamptz, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_delete_complete_v1(name, uuid, uuid, uuid, bigint, text, text, bytea, bytea, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_delete_complete_v1(name, uuid, uuid, uuid, bigint, text, text, bytea, bytea, timestamptz, timestamptz);

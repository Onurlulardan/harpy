-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_unknown_cleanup_defer_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 uuid,
    arg_6 bigint,
    arg_7 uuid,
    arg_8 uuid,
    arg_9 bytea,
    arg_10 bigint,
    arg_11 uuid,
    arg_12 timestamptz
)
RETURNS timestamptz
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT pg_catalog.clock_timestamp();
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_dashboard_export_unknown_cleanup_defer_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, bigint, uuid, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_unknown_cleanup_defer_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, bigint, uuid, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_unknown_cleanup_defer_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, bigint, uuid, timestamptz);

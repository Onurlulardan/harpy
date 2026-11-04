-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_dashboard_export_platform_effect_start_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 uuid,
    arg_6 bigint,
    arg_7 uuid,
    arg_8 uuid,
    arg_9 bytea,
    arg_10 text,
    arg_11 text,
    arg_12 bytea,
    arg_13 bytea,
    arg_14 bytea,
    arg_15 bytea,
    arg_16 bytea,
    arg_17 bigint,
    arg_18 bigint,
    arg_19 bigint,
    arg_20 text,
    arg_21 text,
    arg_22 timestamptz
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

REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_effect_start_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, text, text, bytea, bytea, bytea, bytea, bytea, bigint, bigint, bigint, text, text, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_effect_start_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, text, text, bytea, bytea, bytea, bytea, bytea, bigint, bigint, bigint, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_platform_effect_start_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, text, text, bytea, bytea, bytea, bytea, bytea, bigint, bigint, bigint, text, text, timestamptz);

-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_export_platform_succeed_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 uuid,
    arg_6 bigint,
    arg_7 uuid,
    arg_8 uuid,
    arg_9 bytea,
    arg_10 uuid,
    arg_11 bytea,
    arg_12 text,
    arg_13 text,
    arg_14 timestamptz,
    arg_15 timestamptz
)
RETURNS text
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT ''::text;
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_succeed_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, uuid, bytea, text, text, timestamptz, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_succeed_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, uuid, bytea, text, text, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_platform_succeed_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, uuid, bytea, text, text, timestamptz, timestamptz);

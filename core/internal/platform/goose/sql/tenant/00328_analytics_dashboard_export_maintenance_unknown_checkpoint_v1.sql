-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_export_maintenance_unknown_checkpoint_v1(
    arg_1 name,
    arg_2 text,
    arg_3 uuid,
    arg_4 name,
    arg_5 uuid,
    arg_6 text,
    arg_7 bigint,
    arg_8 bigint,
    arg_9 uuid,
    arg_10 uuid,
    arg_11 bigint,
    arg_12 bigint,
    arg_13 uuid,
    arg_14 uuid,
    arg_15 bytea,
    arg_16 timestamptz,
    arg_17 bigint,
    arg_18 uuid,
    arg_19 bigint,
    arg_20 timestamptz
)
RETURNS timestamptz
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_maintenance_unknown_checkpoint_v1(name, text, uuid, name, uuid, text, bigint, bigint, uuid, uuid, bigint, bigint, uuid, uuid, bytea, timestamptz, bigint, uuid, bigint, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_maintenance_unknown_checkpoint_v1(name, text, uuid, name, uuid, text, bigint, bigint, uuid, uuid, bigint, bigint, uuid, uuid, bytea, timestamptz, bigint, uuid, bigint, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_maintenance_unknown_checkpoint_v1(name, text, uuid, name, uuid, text, bigint, bigint, uuid, uuid, bigint, bigint, uuid, uuid, bytea, timestamptz, bigint, uuid, bigint, timestamptz);

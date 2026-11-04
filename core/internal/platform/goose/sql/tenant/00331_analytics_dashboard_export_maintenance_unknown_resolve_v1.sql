-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_export_maintenance_unknown_resolve_v1(
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
    arg_20 uuid,
    arg_21 uuid,
    arg_22 bytea,
    arg_23 text,
    arg_24 bytea,
    arg_25 text,
    arg_26 bytea,
    arg_27 timestamptz,
    arg_28 text,
    arg_29 text,
    arg_30 text,
    arg_31 bytea,
    arg_32 timestamptz,
    arg_33 timestamptz,
    arg_34 text,
    arg_35 bytea,
    arg_36 bytea,
    arg_37 timestamptz,
    arg_38 timestamptz
)
RETURNS TABLE (
    state text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text;
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_maintenance_unknown_resolve_v1(name, text, uuid, name, uuid, text, bigint, bigint, uuid, uuid, bigint, bigint, uuid, uuid, bytea, timestamptz, bigint, uuid, bigint, uuid, uuid, bytea, text, bytea, text, bytea, timestamptz, text, text, text, bytea, timestamptz, timestamptz, text, bytea, bytea, timestamptz, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_maintenance_unknown_resolve_v1(name, text, uuid, name, uuid, text, bigint, bigint, uuid, uuid, bigint, bigint, uuid, uuid, bytea, timestamptz, bigint, uuid, bigint, uuid, uuid, bytea, text, bytea, text, bytea, timestamptz, text, text, text, bytea, timestamptz, timestamptz, text, bytea, bytea, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_maintenance_unknown_resolve_v1(name, text, uuid, name, uuid, text, bigint, bigint, uuid, uuid, bigint, bigint, uuid, uuid, bytea, timestamptz, bigint, uuid, bigint, uuid, uuid, bytea, text, bytea, text, bytea, timestamptz, text, text, text, bytea, timestamptz, timestamptz, text, bytea, bytea, timestamptz, timestamptz);

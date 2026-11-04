-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_export_unknown_resolve_v1(
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
    arg_12 uuid,
    arg_13 text,
    arg_14 bytea,
    arg_15 text,
    arg_16 text,
    arg_17 text,
    arg_18 bytea,
    arg_19 timestamptz,
    arg_20 timestamptz,
    arg_21 text,
    arg_22 bytea,
    arg_23 bytea,
    arg_24 timestamptz,
    arg_25 timestamptz
)
RETURNS text
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT ''::text;
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_unknown_resolve_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, bigint, uuid, uuid, text, bytea, text, text, text, bytea, timestamptz, timestamptz, text, bytea, bytea, timestamptz, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_unknown_resolve_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, bigint, uuid, uuid, text, bytea, text, text, text, bytea, timestamptz, timestamptz, text, bytea, bytea, timestamptz, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_unknown_resolve_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, uuid, bytea, bigint, uuid, uuid, text, bytea, text, text, text, bytea, timestamptz, timestamptz, text, bytea, bytea, timestamptz, timestamptz);

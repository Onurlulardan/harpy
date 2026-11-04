-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_export_platform_spool_active_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 bigint,
    arg_5 uuid
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_platform_spool_active_v1(name, uuid, uuid, bigint, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_platform_spool_active_v1(name, uuid, uuid, bigint, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_platform_spool_active_v1(name, uuid, uuid, bigint, uuid);

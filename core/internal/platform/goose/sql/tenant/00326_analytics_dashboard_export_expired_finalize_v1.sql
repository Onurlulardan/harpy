-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION analytics_dashboard_export_expired_finalize_v1(
    arg_1 name,
    arg_2 uuid,
    arg_3 uuid,
    arg_4 uuid,
    arg_5 uuid,
    arg_6 bigint,
    arg_7 uuid,
    arg_8 text,
    arg_9 text,
    arg_10 timestamptz
)
RETURNS TABLE (
    state text,
    finalized_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        ''::text,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION analytics_dashboard_export_expired_finalize_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, text, text, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_dashboard_export_expired_finalize_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_dashboard_export_expired_finalize_v1(name, uuid, uuid, uuid, uuid, bigint, uuid, text, text, timestamptz);

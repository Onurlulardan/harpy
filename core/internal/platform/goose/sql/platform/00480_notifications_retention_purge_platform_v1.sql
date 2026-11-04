-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION platform.notifications_retention_purge_platform_v1(
    arg_1 uuid,
    arg_2 timestamptz,
    arg_3 integer
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION platform.notifications_retention_purge_platform_v1(uuid, timestamptz, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.notifications_retention_purge_platform_v1(uuid, timestamptz, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.notifications_retention_purge_platform_v1(uuid, timestamptz, integer);

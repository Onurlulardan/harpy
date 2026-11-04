-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION notifications_retention_purge_attempts_v1(
    arg_1 timestamptz,
    arg_2 integer
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT 0::bigint;
$$;

REVOKE ALL ON FUNCTION notifications_retention_purge_attempts_v1(timestamptz, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION notifications_retention_purge_attempts_v1(timestamptz, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION notifications_retention_purge_attempts_v1(timestamptz, integer);

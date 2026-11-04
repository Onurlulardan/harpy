-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION platform.support_bug_stats_v1(
    arg_1 uuid,
    arg_2 bigint,
    arg_3 bigint,
    arg_4 boolean,
    arg_5 uuid
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT NULL::bytea;
$$;

REVOKE ALL ON FUNCTION platform.support_bug_stats_v1(uuid, bigint, bigint, boolean, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.support_bug_stats_v1(uuid, bigint, bigint, boolean, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.support_bug_stats_v1(uuid, bigint, bigint, boolean, uuid);

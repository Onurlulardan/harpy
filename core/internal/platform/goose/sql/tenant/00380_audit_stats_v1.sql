-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION audit_stats_v1(
    arg_1 timestamptz
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT NULL::bytea;
$$;

REVOKE ALL ON FUNCTION audit_stats_v1(timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION audit_stats_v1(timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION audit_stats_v1(timestamptz);

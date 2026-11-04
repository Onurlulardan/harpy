-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION notifications_retention_purge_deliveries_v1(
    arg_1 timestamptz,
    arg_2 integer
)
RETURNS bigint
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT 0::bigint;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_retention_purge_deliveries_v1(timestamptz, integer) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION notifications_retention_purge_deliveries_v1(timestamptz, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION notifications_retention_purge_deliveries_v1(timestamptz, integer);

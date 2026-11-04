-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION platform.usage_retention_prune_v1(
    arg_1 uuid,
    arg_2 uuid,
    arg_3 timestamptz,
    arg_4 integer,
    arg_5 timestamptz
)
RETURNS integer
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT 0;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.usage_retention_prune_v1(uuid, uuid, timestamptz, integer, timestamptz) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.usage_retention_prune_v1(uuid, uuid, timestamptz, integer, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.usage_retention_prune_v1(uuid, uuid, timestamptz, integer, timestamptz);

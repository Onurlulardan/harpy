-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_query_root_by_version_v1(
    arg_1 name,
    arg_2 uuid
)
RETURNS uuid
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT '00000000-0000-0000-0000-000000000000'::uuid;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_query_root_by_version_v1(name, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_query_root_by_version_v1(name, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_query_root_by_version_v1(name, uuid);

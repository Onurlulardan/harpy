-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION analytics_resource_manifest_v1(
    arg_1 name,
    arg_2 text,
    arg_3 uuid
)
RETURNS bytea
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT NULL::bytea;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_resource_manifest_v1(name, text, uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION analytics_resource_manifest_v1(name, text, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION analytics_resource_manifest_v1(name, text, uuid);

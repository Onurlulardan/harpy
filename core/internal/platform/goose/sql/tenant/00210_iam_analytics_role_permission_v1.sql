-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_analytics_role_permission_v1(
    in_schema_name name,
    in_tenant_id uuid,
    in_role_id uuid,
    in_resource text,
    in_action text
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_analytics_role_permission_v1(name, uuid, uuid, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_analytics_role_permission_v1(name, uuid, uuid, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_analytics_role_permission_v1(name, uuid, uuid, text, text);

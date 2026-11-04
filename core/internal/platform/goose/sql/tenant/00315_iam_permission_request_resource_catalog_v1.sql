-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_permission_request_resource_catalog_v1(
    in_schema_name name,
    in_actor_id uuid,
    in_kind text,
    in_search text,
    in_after uuid,
    in_arg bigint,
    in_action text
)
RETURNS TABLE (
    id uuid,
    key text,
    display_name text,
    description text,
    kind text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text,
        in_kind
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_permission_request_resource_catalog_v1(name, uuid, text, text, uuid, bigint, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_permission_request_resource_catalog_v1(name, uuid, text, text, uuid, bigint, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_permission_request_resource_catalog_v1(name, uuid, text, text, uuid, bigint, text);

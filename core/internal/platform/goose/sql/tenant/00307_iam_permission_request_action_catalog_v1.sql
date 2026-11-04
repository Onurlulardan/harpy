-- +goose Up
CREATE FUNCTION iam_permission_request_action_catalog_v1(
    in_schema_name name,
    in_actor_id uuid,
    in_search text,
    in_after uuid,
    in_arg bigint
)
RETURNS TABLE (
    id uuid,
    key text,
    display_name text,
    description text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text
    ;
$$;

REVOKE ALL ON FUNCTION iam_permission_request_action_catalog_v1(name, uuid, text, uuid, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_permission_request_action_catalog_v1(name, uuid, text, uuid, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_permission_request_action_catalog_v1(name, uuid, text, uuid, bigint);

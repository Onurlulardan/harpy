-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_rbac_permission_get_v1(
    in_schema_name name,
    in_id uuid,
    in_actor_id uuid,
    in_required_permission_id uuid
)
RETURNS TABLE (
    id uuid,
    match_kind text,
    resource_id uuid,
    action_id uuid,
    protected boolean,
    state text,
    revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_id,
        ''::text,
        NULL::uuid,
        NULL::uuid,
        false,
        ''::text,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_rbac_permission_get_v1(name, uuid, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rbac_permission_get_v1(name, uuid, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rbac_permission_get_v1(name, uuid, uuid, uuid);

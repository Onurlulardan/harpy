-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_rbac_effective_permissions_v1(
    in_schema_name name,
    in_target text,
    in_target_id uuid,
    in_limit integer,
    in_actor_id uuid,
    in_required_permission_id uuid
)
RETURNS TABLE (
    permission_id uuid,
    match_kind text,
    resource_id uuid,
    resource_key text,
    action_id uuid,
    action_key text,
    source text,
    source_id uuid
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        NULL::uuid,
        ''::text,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid
    ;
$$;

REVOKE ALL ON FUNCTION iam_rbac_effective_permissions_v1(name, text, uuid, integer, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rbac_effective_permissions_v1(name, text, uuid, integer, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rbac_effective_permissions_v1(name, text, uuid, integer, uuid, uuid);

-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_rbac_role_binding_list_v1(
    in_schema_name name,
    in_subject_id uuid,
    in_state text,
    in_limit integer,
    in_actor_id uuid,
    in_required_permission_id uuid
)
RETURNS TABLE (
    id uuid,
    subject_id uuid,
    permission_id uuid,
    state text,
    revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_subject_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_state,
        0::bigint
    ;
$$;

REVOKE ALL ON FUNCTION iam_rbac_role_binding_list_v1(name, uuid, text, integer, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rbac_role_binding_list_v1(name, uuid, text, integer, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rbac_role_binding_list_v1(name, uuid, text, integer, uuid, uuid);

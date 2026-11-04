-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
CREATE FUNCTION iam_rbac_assignment_read_v1(
    in_schema_name name,
    in_id uuid,
    in_actor_id uuid,
    in_required_permission_id uuid
)
RETURNS TABLE (
    id uuid,
    resource_id uuid,
    target text,
    target_id uuid,
    state text,
    revision bigint,
    created_at timestamptz,
    updated_at timestamptz,
    items_json bytea
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        ''::bytea
    ;
$$;

REVOKE ALL ON FUNCTION iam_rbac_assignment_read_v1(name, uuid, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rbac_assignment_read_v1(name, uuid, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rbac_assignment_read_v1(name, uuid, uuid, uuid);

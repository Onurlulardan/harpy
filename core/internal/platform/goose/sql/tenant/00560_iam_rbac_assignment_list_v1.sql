-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_assignment_list_v1(
    in_schema_name name,
    in_state text,
    in_value uuid,
    in_limit integer,
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
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        in_state,
        0::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        ''::bytea
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_assignment_list_v1(name, text, uuid, integer, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rbac_assignment_list_v1(name, text, uuid, integer, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rbac_assignment_list_v1(name, text, uuid, integer, uuid, uuid);

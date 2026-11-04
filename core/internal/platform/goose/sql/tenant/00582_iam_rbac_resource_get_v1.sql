-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_resource_get_v1(
    in_schema_name name,
    in_id uuid,
    in_actor_id uuid,
    in_required_permission_id uuid
)
RETURNS TABLE (
    id uuid,
    key text,
    owner_domain text,
    kind text,
    display_name text,
    description text,
    protected boolean,
    state text,
    revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_id,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        NULL::text,
        false,
        ''::text,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_resource_get_v1(name, uuid, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rbac_resource_get_v1(name, uuid, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rbac_resource_get_v1(name, uuid, uuid, uuid);

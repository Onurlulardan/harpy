-- +goose Up
-- BELIRSIZ: caller does not define routine behavior.
-- +goose StatementBegin
CREATE FUNCTION iam_rbac_protected_admin_get_v1(
    in_schema_name name,
    in_id uuid,
    in_actor_id uuid,
    in_required_permission_id uuid
)
RETURNS TABLE (
    role_id uuid,
    permission_id uuid,
    binding_id uuid,
    minimum_active_assignments bigint,
    authorization_epoch bigint,
    revision bigint
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        0::bigint,
        0::bigint
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_rbac_protected_admin_get_v1(name, uuid, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_rbac_protected_admin_get_v1(name, uuid, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_rbac_protected_admin_get_v1(name, uuid, uuid, uuid);

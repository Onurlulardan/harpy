-- +goose Up
CREATE FUNCTION platform.iam_post_restore_binding_v1(
    in_environment_id uuid,
    in_restore_operation_id uuid
)
RETURNS TABLE (
    tenant_id uuid,
    owner_user_id uuid,
    anchor_id uuid,
    new_database_system_id uuid,
    security_epoch bytea,
    tenant_slug text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::bytea,
        ''::text
    ;
$$;

REVOKE ALL ON FUNCTION platform.iam_post_restore_binding_v1(uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_post_restore_binding_v1(uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_post_restore_binding_v1(uuid, uuid);

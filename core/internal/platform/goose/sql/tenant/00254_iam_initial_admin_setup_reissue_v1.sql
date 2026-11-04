-- +goose Up
CREATE FUNCTION iam_initial_admin_setup_reissue_v1(
    in_schema_name name,
    in_operation_id uuid,
    in_reissue_operation_id uuid,
    in_tenant_id uuid
)
RETURNS uuid
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT '00000000-0000-0000-0000-000000000000'::uuid;
$$;

REVOKE ALL ON FUNCTION iam_initial_admin_setup_reissue_v1(name, uuid, uuid, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION iam_initial_admin_setup_reissue_v1(name, uuid, uuid, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION iam_initial_admin_setup_reissue_v1(name, uuid, uuid, uuid);

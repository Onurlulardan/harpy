-- +goose Up
CREATE FUNCTION platform.iam_installation_activation_assert_v2(
    in_operation_id uuid,
    in_environment_id uuid,
    in_tenant_id uuid,
    in_schema_name name
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.iam_installation_activation_assert_v2(uuid, uuid, uuid, name) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.iam_installation_activation_assert_v2(uuid, uuid, uuid, name) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.iam_installation_activation_assert_v2(uuid, uuid, uuid, name);

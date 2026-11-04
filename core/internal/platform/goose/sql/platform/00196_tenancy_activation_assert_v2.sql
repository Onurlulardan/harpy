-- +goose Up
-- BELIRSIZ: activation assertion is not determined by the caller.
CREATE FUNCTION platform.tenancy_activation_assert_v2(
    p_operation_id uuid,
    p_tenant_id uuid,
    p_schema_name name
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT false;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_activation_assert_v2(
    uuid, uuid, name
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_assert_v2(
    uuid, uuid, name
) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_activation_assert_v2(
    uuid, uuid, name
);

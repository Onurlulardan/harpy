-- +goose Up
CREATE FUNCTION platform.tenancy_get_primary_tenant_target_v1()
RETURNS TABLE (
    environment_id uuid,
    deployment_mode text,
    tenant_id uuid,
    purpose text,
    slug text
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        NULL::uuid,
        NULL::text,
        NULL::uuid,
        NULL::text,
        NULL::text
    WHERE false;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_get_primary_tenant_target_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_get_primary_tenant_target_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_get_primary_tenant_target_v1();

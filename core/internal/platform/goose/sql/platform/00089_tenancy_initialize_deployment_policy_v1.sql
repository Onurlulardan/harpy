-- +goose Up
CREATE FUNCTION platform.tenancy_initialize_deployment_policy_v1(
    in_operation_id uuid,
    in_environment_id uuid,
    in_operation_kind text,
    in_deployment_mode text
)
RETURNS TABLE (
    operation_id uuid,
    environment_id uuid,
    deployment_mode text,
    maximum_customer_tenants bigint,
    created_at timestamptz,
    database_now timestamptz,
    replayed boolean
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_operation_id,
        in_environment_id,
        in_deployment_mode,
        NULL::bigint,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp(),
        false;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_initialize_deployment_policy_v1(uuid, uuid, text, text) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_initialize_deployment_policy_v1(uuid, uuid, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_initialize_deployment_policy_v1(uuid, uuid, text, text);

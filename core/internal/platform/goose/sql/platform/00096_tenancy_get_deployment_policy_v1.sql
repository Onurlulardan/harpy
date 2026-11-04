-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_get_deployment_policy_v1()
RETURNS TABLE (
    environment_id uuid,
    deployment_mode text,
    maximum_customer_tenants bigint,
    created_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        policy.environment_id,
        policy.deployment_mode,
        policy.maximum_customer_tenants,
        policy.created_at
    FROM platform.tenancy_deployment_policy AS policy
    WHERE policy.environment_id IS NOT NULL
      AND policy.environment_id <> '00000000-0000-0000-0000-000000000000'::uuid;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_get_deployment_policy_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_get_deployment_policy_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_get_deployment_policy_v1();

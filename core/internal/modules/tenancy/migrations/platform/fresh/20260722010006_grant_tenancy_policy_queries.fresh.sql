-- +goose Up
REVOKE ALL ON FUNCTION
    platform.tenancy_get_deployment_policy_v1()
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.tenancy_get_primary_tenant_target_v1()
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.tenancy_assert_saas_customer_target_v1(uuid, uuid)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.tenancy_assert_saas_customer_offboard_v1(uuid, uuid)
FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.tenancy_list_saas_customer_tenants_v1(uuid)
FROM PUBLIC;

GRANT EXECUTE ON FUNCTION
    platform.tenancy_get_deployment_policy_v1()
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.tenancy_get_primary_tenant_target_v1()
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.tenancy_assert_saas_customer_target_v1(uuid, uuid)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.tenancy_assert_saas_customer_offboard_v1(uuid, uuid)
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION
    platform.tenancy_list_saas_customer_tenants_v1(uuid)
TO forgecrud_runtime;

GRANT USAGE ON SCHEMA platform TO forgecrud_runtime;

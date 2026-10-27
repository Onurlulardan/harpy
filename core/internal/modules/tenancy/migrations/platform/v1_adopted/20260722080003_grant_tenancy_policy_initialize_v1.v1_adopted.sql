-- +goose Up
REVOKE ALL ON FUNCTION platform.tenancy_initialize_deployment_policy_v1(
    uuid,uuid,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_initialize_deployment_policy_v1(
    uuid,uuid,text,text
) TO harpy_runtime;

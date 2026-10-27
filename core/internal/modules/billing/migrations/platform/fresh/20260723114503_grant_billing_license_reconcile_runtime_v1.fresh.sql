-- +goose Up
GRANT EXECUTE ON FUNCTION platform.billing_license_refresh_target_acquire_v1()
TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.billing_license_restore_binding_v1(uuid)
TO forgecrud_runtime;

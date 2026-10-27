-- +goose Up
GRANT EXECUTE ON FUNCTION platform.billing_license_runtime_entitlements_v1(uuid)
TO forgecrud_runtime;


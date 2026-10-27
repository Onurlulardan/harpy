-- +goose Up
GRANT EXECUTE ON FUNCTION
platform.billing_platform_license_refresh_request_v1(
    uuid, uuid, uuid, text, text
)
TO forgecrud_runtime;

GRANT EXECUTE ON FUNCTION
platform.billing_platform_license_refresh_read_v1(uuid, uuid)
TO forgecrud_runtime;

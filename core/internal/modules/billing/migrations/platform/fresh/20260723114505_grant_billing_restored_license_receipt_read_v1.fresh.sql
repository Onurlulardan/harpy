-- +goose Up
GRANT EXECUTE ON FUNCTION
platform.billing_restored_license_receipt_read_v1(uuid, uuid)
TO harpy_runtime;

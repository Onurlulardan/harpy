-- +goose Up
GRANT EXECUTE ON FUNCTION platform.billing_commit_restored_license_v1(
    uuid, uuid, text, uuid, uuid, uuid, uuid, bigint, text, bigint, text,
    bigint, text, text, text, text, timestamptz, timestamptz, timestamptz,
    timestamptz, timestamptz, timestamptz, text, jsonb, jsonb,
    bytea, bytea, jsonb
) TO forgecrud_runtime;


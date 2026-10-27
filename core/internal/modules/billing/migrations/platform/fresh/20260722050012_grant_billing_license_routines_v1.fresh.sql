-- +goose Up
GRANT EXECUTE ON FUNCTION platform.billing_license_refresh_claim_v1(
    uuid, uuid, uuid, uuid, text, jsonb, integer
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.billing_license_refresh_failure_v1(
    uuid, uuid, text
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.billing_license_snapshot_accept_v1(
    uuid, uuid, uuid, uuid, uuid, text, text, text, text, text, bigint,
    timestamptz, timestamptz, timestamptz, timestamptz, timestamptz,
    timestamptz, text, jsonb, jsonb, bytea, bytea, jsonb
) TO harpy_runtime;
GRANT EXECUTE ON FUNCTION platform.billing_publish_license_derived_offering_v1(
    uuid, uuid, uuid, uuid, text, text, text, bigint, text, uuid,
    bigint, text, timestamptz, text, bytea
) TO harpy_runtime;


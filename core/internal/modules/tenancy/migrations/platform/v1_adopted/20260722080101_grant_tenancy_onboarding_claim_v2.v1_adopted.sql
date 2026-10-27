-- +goose Up
REVOKE ALL ON FUNCTION platform.tenancy_onboarding_claim_v2(
    uuid, uuid, text, text, bigint, integer
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_claim_v2(
    uuid, uuid, text, text, bigint, integer
) TO harpy_runtime;

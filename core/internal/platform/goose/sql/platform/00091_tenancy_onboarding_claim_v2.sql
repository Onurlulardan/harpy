-- +goose Up
CREATE FUNCTION platform.tenancy_onboarding_claim_v2(
    in_operation_id uuid,
    in_attempt_id uuid,
    in_lease_owner text,
    in_lease_token_sha256 text,
    in_expected_revision bigint,
    in_lease_seconds integer
)
RETURNS TABLE (
    operation_id uuid,
    attempt_id uuid,
    attempt_number bigint,
    revision bigint,
    lease_until timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_operation_id,
        in_attempt_id,
        0::bigint,
        in_expected_revision,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer);

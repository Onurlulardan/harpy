-- +goose Up
-- +goose StatementBegin
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
AS $fn$
    UPDATE platform.tenancy_onboarding_operations AS operation
    SET attempt_number = operation.attempt_number + 1,
        updated_at = pg_catalog.clock_timestamp()
    WHERE operation.operation_id = in_operation_id
      AND operation.revision = in_expected_revision
    RETURNING operation.operation_id, in_attempt_id, operation.attempt_number,
              operation.revision,
              pg_catalog.clock_timestamp()
                  + (in_lease_seconds * pg_catalog.interval '1 second');
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_claim_v2(uuid, uuid, text, text, bigint, integer);

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_record_v1(
    in_receipt_id uuid,
    in_operation_id uuid,
    in_attempt_id uuid,
    in_kind text,
    in_sha256 text,
    in_evidence_sha256 text,
    in_occurred_at timestamptz,
    in_expected_revision bigint
)
RETURNS TABLE (
    operation_id uuid,
    state text,
    revision bigint,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        in_operation_id,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_record_v1(uuid, uuid, uuid, text, text, text, timestamptz, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_record_v1(uuid, uuid, uuid, text, text, text, timestamptz, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_record_v1(uuid, uuid, uuid, text, text, text, timestamptz, bigint);

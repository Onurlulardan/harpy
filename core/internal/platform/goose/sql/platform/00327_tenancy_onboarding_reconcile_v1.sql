-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_reconcile_v1(
    in_operation_id uuid,
    in_receipt_id uuid,
    in_receipt_sha256 text,
    in_evidence_sha256 text,
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

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_reconcile_v1(uuid, uuid, text, text, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_reconcile_v1(uuid, uuid, text, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_reconcile_v1(uuid, uuid, text, text, bigint);

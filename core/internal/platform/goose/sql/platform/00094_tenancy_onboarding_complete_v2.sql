-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_complete_v2(
    in_operation_id uuid,
    in_receipt_id uuid,
    in_receipt_sha256 text,
    in_evidence_sha256 text,
    in_expected_revision bigint
)
RETURNS boolean
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT false;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_complete_v2(uuid, uuid, text, text, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_complete_v2(uuid, uuid, text, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_complete_v2(uuid, uuid, text, text, bigint);

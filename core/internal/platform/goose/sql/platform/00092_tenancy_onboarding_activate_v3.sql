-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_activate_v3(
    in_operation_id uuid,
    in_lease_token_sha256 text,
    in_expected_revision bigint
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    operation_kind text,
    step_plan_sha256 text,
    migration_receipt_id uuid,
    migration_receipt_sha256 text,
    authority_ack_sha256 text,
    revision bigint,
    database_now timestamptz,
    lease_until timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        operation.operation_id,
        operation.tenant_id,
        operation.kind,
        operation.step_plan_sha256,
        pg_catalog.gen_random_uuid(),
        pg_catalog.md5(operation.operation_id::text) || pg_catalog.md5(operation.tenant_id::text),
        NULL::text,
        operation.revision,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp() + pg_catalog.interval '1 minute'
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id = in_operation_id
      AND operation.revision = in_expected_revision;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint);

-- +goose Up
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
AS $$
    SELECT
        in_operation_id,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        NULL::text,
        in_expected_revision,
        pg_catalog.clock_timestamp(),
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_activate_v3(uuid, text, bigint);

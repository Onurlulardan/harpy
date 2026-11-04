-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_provision_claim_v1(
    in_attempt_id uuid,
    in_lease_owner text,
    in_lease_token_sha256 text,
    in_seconds integer
)
RETURNS TABLE (
    operation_id uuid,
    environment_id uuid,
    operation_kind text,
    idempotency_key text,
    request_sha text,
    manifest_sha text,
    step_plan_sha text,
    deployment_mode text,
    tenant_id uuid,
    schema_name text,
    binding_purpose text,
    first_user_id uuid,
    root_organization_id uuid,
    protected_admin_id uuid,
    instance_owner uuid,
    offering_version_id uuid,
    offering_revision bigint,
    offering_sha text,
    offering_effective_at timestamptz,
    source_sequence bigint,
    source_sha text,
    recovery_generation bigint,
    recovery_sha text,
    registry_binding_id uuid,
    authority_kind text,
    authority_scope_id uuid,
    authority_generation bigint,
    previous_head text,
    authority_state text,
    authority_proposal text,
    created_at timestamptz,
    customer_name text,
    tenant_slug text,
    admin_display_name text,
    admin_email text,
    state text,
    revision bigint,
    lease_until timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $fn$
    SELECT
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        NULL::uuid,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        ''::text,
        pg_catalog.clock_timestamp(),
        NULL::bigint,
        NULL::text,
        NULL::bigint,
        NULL::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        ''::text,
        '00000000-0000-0000-0000-000000000000'::uuid,
        0::bigint,
        NULL::text,
        ''::text,
        ''::text,
        pg_catalog.clock_timestamp(),
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp()
    ;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_provision_claim_v1(uuid, text, text, integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_provision_claim_v1(uuid, text, text, integer) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_provision_claim_v1(uuid, text, text, integer);

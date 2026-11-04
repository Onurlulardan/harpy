-- +goose Up
CREATE FUNCTION platform.tenancy_onboarding_accept_v2(
    in_operation_id uuid,
    in_environment_id uuid,
    in_kind text,
    in_idempotency_key text,
    in_request_sha256 text,
    in_manifest_sha256 text,
    in_step_plan_sha256 text,
    in_deployment_mode text,
    in_tenant_id uuid,
    in_schema_name name,
    in_binding_purpose text,
    in_first_user_id uuid,
    in_root_organization_id uuid,
    in_protected_admin_binding_id uuid,
    in_instance_owner_binding_id uuid,
    in_offering_version_id uuid,
    in_offering_revision bigint,
    in_offering_sha256 text,
    in_offering_effective_at timestamptz,
    in_source_sequence bigint,
    in_source_sha256 text,
    in_recovery_generation bigint,
    in_trust_bundle_sha256 text,
    in_registry_binding_id uuid,
    in_scope_kind text,
    in_scope_id uuid,
    in_authority_generation bigint,
    in_previous_head_sha256 text,
    in_state_sha256 text,
    in_proposal_sha256 text,
    in_created_at timestamptz
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
AS $$
    SELECT
        in_operation_id,
        ''::text,
        0::bigint,
        pg_catalog.clock_timestamp();
$$;

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_onboarding_accept_v2(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz);

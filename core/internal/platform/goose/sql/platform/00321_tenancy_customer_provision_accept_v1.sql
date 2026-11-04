-- +goose Up
CREATE FUNCTION platform.tenancy_customer_provision_accept_v1(
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
    in_version_id uuid,
    in_revision bigint,
    in_sha256 text,
    in_effective_at timestamptz,
    in_arg bigint,
    in_null_string text,
    in_arg_2 bigint,
    in_null_string_2 text,
    in_registry_binding_id uuid,
    in_scope_kind text,
    in_scope_id uuid,
    in_generation bigint,
    in_previous_head_sha256 text,
    in_state_sha256 text,
    in_proposal_sha256 text,
    in_created_at timestamptz,
    in_control_tenant_id uuid,
    in_requested_by_subject_id uuid,
    in_name text,
    in_slug text,
    in_display_name text,
    in_email text,
    in_job_id uuid
)
RETURNS TABLE (
    operation_id uuid,
    tenant_id uuid,
    requested_by_subject_id uuid,
    state text,
    setup_delivery_state text,
    failure_code text,
    failure_retryable boolean,
    correlation_id text,
    created_at timestamptz,
    updated_at timestamptz
)
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        in_operation_id,
        in_tenant_id,
        in_requested_by_subject_id,
        ''::text,
        ''::text,
        NULL::text,
        NULL::boolean,
        NULL::text,
        in_created_at,
        pg_catalog.clock_timestamp()
    ;
$$;

REVOKE ALL ON FUNCTION platform.tenancy_customer_provision_accept_v1(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz, uuid, uuid, text, text, text, text, uuid) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.tenancy_customer_provision_accept_v1(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz, uuid, uuid, text, text, text, text, uuid) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_customer_provision_accept_v1(uuid, uuid, text, text, text, text, text, text, uuid, name, text, uuid, uuid, uuid, uuid, uuid, bigint, text, timestamptz, bigint, text, bigint, text, uuid, text, uuid, bigint, text, text, text, timestamptz, uuid, uuid, text, text, text, text, uuid);

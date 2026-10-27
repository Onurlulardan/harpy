-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_customer_provision_accept_v1(
    p_operation_id uuid, p_environment_id uuid, p_operation_kind text,
    p_idempotency_key text, p_request_sha256 text, p_manifest_sha256 text,
    p_step_plan_sha256 text, p_deployment_mode text, p_tenant_id uuid,
    p_schema_name name, p_binding_purpose text, p_first_user_id uuid,
    p_root_organization_id uuid, p_protected_admin_binding_id uuid,
    p_instance_owner_binding_id uuid, p_offering_version_id uuid,
    p_offering_revision bigint, p_offering_sha256 text,
    p_offering_effective_at timestamptz, p_offering_source_sequence bigint,
    p_offering_source_sha256 text, p_recovery_signing_generation bigint,
    p_recovery_trust_bundle_sha256 text, p_registry_binding_id uuid,
    p_authority_scope_kind text, p_authority_scope_id uuid,
    p_authority_generation bigint, p_authority_previous_head_sha256 text,
    p_authority_state_sha256 text, p_authority_proposal_sha256 text,
    p_created_at timestamptz, p_control_tenant_id uuid,
    p_requested_by_subject_id uuid, p_customer_name text, p_tenant_slug text,
    p_first_admin_display_name text, p_first_admin_email text, p_job_id uuid
)
RETURNS TABLE (
    operation_id uuid, tenant_id uuid, requested_by_subject_id uuid,
    state text, setup_delivery_state text, failure_code text,
    failure_retryable boolean, correlation_id text,
    created_at timestamptz, updated_at timestamptz
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_ready platform.tenancy_tenant_migration_executor_readiness%ROWTYPE;
    v_request platform.tenancy_customer_provisioning_requests%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_operation_kind <> 'saas_customer_tenant_provision'
       OR p_deployment_mode <> 'saas' OR p_binding_purpose <> 'saas_customer'
       OR p_instance_owner_binding_id IS NOT NULL
       OR p_requested_by_subject_id IS NULL OR p_control_tenant_id IS NULL
       OR p_job_id IS NULL OR p_request_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='customer provisioning acceptance is invalid';
    END IF;
    PERFORM 1
    FROM platform.tenancy_deployment_policy AS policy
    JOIN platform.tenancy_tenant_bindings AS binding
      ON binding.environment_id=policy.environment_id
     AND binding.purpose='saas_control'
    JOIN platform.tenants AS tenant ON tenant.id=binding.tenant_id
    WHERE policy.environment_id=p_environment_id
      AND policy.deployment_mode='saas'
      AND binding.tenant_id=p_control_tenant_id
      AND tenant.status='ACTIVE' AND tenant.deleted_at IS NULL
    FOR SHARE OF policy,binding,tenant;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='active SaaS control tenant is required';
    END IF;
    SELECT ready.* INTO v_ready
    FROM platform.tenancy_tenant_migration_executor_readiness AS ready
    WHERE ready.environment_id=p_environment_id
      AND ready.deployment_mode='saas' AND ready.state='active'
      AND ready.ready_until > pg_catalog.clock_timestamp()
    ORDER BY ready.ready_until DESC,ready.release_id DESC
    LIMIT 1 FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='tenant migration executor is not claim ready';
    END IF;
    PERFORM 1 FROM platform.tenancy_onboarding_accept_v1(
        p_operation_id,p_environment_id,p_operation_kind,p_idempotency_key,
        p_request_sha256,p_manifest_sha256,p_step_plan_sha256,p_deployment_mode,
        p_tenant_id,p_schema_name,p_binding_purpose,p_first_user_id,
        p_root_organization_id,p_protected_admin_binding_id,
        p_instance_owner_binding_id,p_offering_version_id,p_offering_revision,
        p_offering_sha256,p_offering_effective_at,p_offering_source_sequence,
        p_offering_source_sha256,p_recovery_signing_generation,
        p_recovery_trust_bundle_sha256,p_registry_binding_id,
        p_authority_scope_kind,p_authority_scope_id,p_authority_generation,
        p_authority_previous_head_sha256,p_authority_state_sha256,
        p_authority_proposal_sha256,p_created_at
    );
    INSERT INTO platform.tenancy_customer_provisioning_requests (
        operation_id,control_tenant_id,requested_by_subject_id,customer_name,
        tenant_slug,first_admin_display_name,first_admin_email,
        request_sha256,created_at
    ) VALUES (
        p_operation_id,p_control_tenant_id,p_requested_by_subject_id,p_customer_name,
        p_tenant_slug,p_first_admin_display_name,p_first_admin_email,
        p_request_sha256,p_created_at
    ) ON CONFLICT (operation_id) DO NOTHING;
    SELECT request.* INTO STRICT v_request
    FROM platform.tenancy_customer_provisioning_requests AS request
    WHERE request.operation_id=p_operation_id FOR SHARE;
    IF v_request.control_tenant_id IS DISTINCT FROM p_control_tenant_id
       OR v_request.requested_by_subject_id IS DISTINCT FROM p_requested_by_subject_id
       OR v_request.customer_name IS DISTINCT FROM p_customer_name
       OR v_request.tenant_slug IS DISTINCT FROM p_tenant_slug
       OR v_request.first_admin_display_name IS DISTINCT FROM p_first_admin_display_name
       OR v_request.first_admin_email IS DISTINCT FROM p_first_admin_email
       OR v_request.request_sha256 IS DISTINCT FROM p_request_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='customer provisioning idempotency conflicts';
    END IF;
    INSERT INTO platform.tenancy_customer_provisioning_outcomes (
        operation_id,setup_delivery_state,revision,updated_at
    ) VALUES (p_operation_id,'pending',1,p_created_at)
    ON CONFLICT (operation_id) DO NOTHING;
    PERFORM 1 FROM platform.tenancy_tenant_migration_enqueue_v1(
        p_job_id,p_operation_id,v_ready.release_id,v_ready.runtime_generation,
        v_ready.database_generation,v_ready.migration_manifest_sha256
    );
    RETURN QUERY SELECT operation.operation_id,operation.tenant_id,
        request.requested_by_subject_id,operation.state,outcome.setup_delivery_state,
        outcome.failure_code,outcome.failure_retryable,outcome.correlation_id::text,
        operation.created_at,operation.updated_at
    FROM platform.tenancy_onboarding_operations AS operation
    JOIN platform.tenancy_customer_provisioning_requests AS request
      ON request.operation_id=operation.operation_id
    JOIN platform.tenancy_customer_provisioning_outcomes AS outcome
      ON outcome.operation_id=operation.operation_id
    WHERE operation.operation_id=p_operation_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_customer_provision_accept_v1(
    uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,
    uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,
    text,text,text,timestamptz,uuid,uuid,text,text,text,text,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_customer_provision_accept_v1(
    uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,uuid,
    uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,uuid,bigint,
    text,text,text,timestamptz,uuid,uuid,text,text,text,text,uuid
) TO harpy_runtime;

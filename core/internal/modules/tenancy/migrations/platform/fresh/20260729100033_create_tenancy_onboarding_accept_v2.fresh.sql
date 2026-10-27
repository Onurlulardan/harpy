-- +goose Up
-- Migration stream version: 20260729100033.
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_accept_v2(
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
    p_created_at timestamptz
)
RETURNS TABLE (
    operation_id uuid, state text, revision bigint, updated_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.tenancy_onboarding_operations%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'onboarding accept denied';
    END IF;
    PERFORM 1
    FROM platform.tenancy_deployment_policy AS policy
    WHERE policy.environment_id = p_environment_id
      AND policy.deployment_mode = p_deployment_mode
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'deployment policy mismatch';
    END IF;
    SELECT * INTO v_existing
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.environment_id = p_environment_id
      AND operation.operation_kind = p_operation_kind
      AND operation.idempotency_key = p_idempotency_key
    FOR UPDATE;
    IF FOUND THEN
        IF v_existing.operation_id IS DISTINCT FROM p_operation_id
           OR v_existing.request_sha256 IS DISTINCT FROM p_request_sha256
           OR v_existing.manifest_sha256 IS DISTINCT FROM p_manifest_sha256
           OR v_existing.step_plan_sha256
                IS DISTINCT FROM p_step_plan_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'onboarding idempotency conflict';
        END IF;
        RETURN QUERY SELECT v_existing.operation_id, v_existing.state,
            v_existing.revision, v_existing.updated_at;
        RETURN;
    END IF;
    INSERT INTO platform.tenancy_onboarding_operations (
        operation_id, environment_id, operation_kind, idempotency_key,
        request_sha256, manifest_sha256, step_plan_sha256, deployment_mode,
        tenant_id, schema_name, binding_purpose, first_user_id,
        root_organization_id, protected_admin_binding_id,
        instance_owner_binding_id, offering_version_id, offering_revision,
        offering_sha256, offering_effective_at, offering_source_sequence,
        offering_source_sha256, recovery_signing_generation,
        recovery_trust_bundle_sha256, registry_binding_id,
        authority_scope_kind, authority_scope_id, authority_generation,
        authority_previous_head_sha256, authority_state_sha256,
        authority_proposal_sha256, created_at, updated_at
    ) VALUES (
        p_operation_id, p_environment_id, p_operation_kind,
        p_idempotency_key, p_request_sha256, p_manifest_sha256,
        p_step_plan_sha256, p_deployment_mode, p_tenant_id, p_schema_name,
        p_binding_purpose, p_first_user_id, p_root_organization_id,
        p_protected_admin_binding_id, p_instance_owner_binding_id,
        p_offering_version_id, p_offering_revision, p_offering_sha256,
        p_offering_effective_at, p_offering_source_sequence,
        p_offering_source_sha256, p_recovery_signing_generation,
        p_recovery_trust_bundle_sha256, p_registry_binding_id,
        p_authority_scope_kind, p_authority_scope_id,
        p_authority_generation, p_authority_previous_head_sha256,
        p_authority_state_sha256, p_authority_proposal_sha256,
        p_created_at, p_created_at
    );
    IF p_operation_kind = 'environment_initialize' THEN
        INSERT INTO platform.tenancy_onboarding_steps (
            operation_id, step_ordinal, command_key
        ) VALUES
            (p_operation_id, 1, 'tenancy.activate_tenant'),
            (p_operation_id, 2, 'iam.initialize_tenant_identity'),
            (p_operation_id, 3, 'billing.assign_tenant_base_offering'),
            (p_operation_id, 4, 'billing.open_tenant_usage'),
            (p_operation_id, 5, 'systemresources.initialize_tenant_baseline'),
            (p_operation_id, 6, 'platformops.initialize_recovery_signing_identity'),
            (p_operation_id, 7, 'platformauthority.bind_owner'),
            (p_operation_id, 8, 'platformauthority.pin_protected_authority_head'),
            (p_operation_id, 9, 'activation.complete');
    ELSE
        INSERT INTO platform.tenancy_onboarding_steps (
            operation_id, step_ordinal, command_key
        ) VALUES
            (p_operation_id, 1, 'tenancy.activate_tenant'),
            (p_operation_id, 2, 'iam.initialize_tenant_identity'),
            (p_operation_id, 3, 'billing.assign_tenant_base_offering'),
            (p_operation_id, 4, 'billing.open_tenant_usage'),
            (p_operation_id, 5, 'systemresources.initialize_tenant_baseline'),
            (p_operation_id, 6, 'platformauthority.pin_protected_authority_head'),
            (p_operation_id, 7, 'activation.complete');
    END IF;
    RETURN QUERY SELECT p_operation_id, 'accepted'::text,
        1::bigint, p_created_at;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_accept_v2(
    uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,
    uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,
    uuid,bigint,text,text,text,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_accept_v2(
    uuid,uuid,text,text,text,text,text,text,uuid,name,text,uuid,uuid,uuid,
    uuid,uuid,bigint,text,timestamptz,bigint,text,bigint,text,uuid,text,
    uuid,bigint,text,text,text,timestamptz
) TO forgecrud_runtime;

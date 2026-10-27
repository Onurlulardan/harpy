-- +goose Up
-- Migration stream version: 20260729100031.
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_complete_v2(
    p_operation_id uuid,
    p_receipt_id uuid,
    p_receipt_sha256 text,
    p_evidence_sha256 text,
    p_expected_revision bigint
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.tenancy_onboarding_operations%ROWTYPE;
    v_artifact_id uuid;
BEGIN
    SELECT * INTO v_operation
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id = p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR NOT platform.tenancy_activation_step_assert_v1(
        p_operation_id, 'activation.complete'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant activation completion capability required';
    END IF;
    IF v_operation.operation_kind = 'environment_initialize'
       AND v_operation.deployment_mode = 'onpremise' THEN
        IF v_operation.authority_generation IS DISTINCT FROM 0
           OR v_operation.authority_scope_kind IS DISTINCT FROM 'environment'
           OR NOT platform.tenancy_activation_step_consume_v1(
               p_operation_id,
               'platformauthority.pin_protected_authority_head'
           ) THEN
            RAISE EXCEPTION USING ERRCODE = '42501',
                MESSAGE = 'on-premise authority no-op capability required';
        END IF;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.tenancy_activation_step_capabilities AS capability
        WHERE capability.backend_pid = pg_catalog.pg_backend_pid()
          AND capability.transaction_id = pg_catalog.pg_current_xact_id()
          AND capability.operation_id = p_operation_id
          AND capability.command_key <> 'activation.complete'
    ) OR EXISTS (
        SELECT 1
        FROM platform.tenancy_activation_command_capabilities AS capability
        WHERE capability.backend_pid = pg_catalog.pg_backend_pid()
          AND capability.transaction_id = pg_catalog.pg_current_xact_id()
          AND capability.operation_id = p_operation_id
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'tenant activation has incomplete command capabilities';
    END IF;
    IF NOT platform.platform_async_installation_activation_evidence_assert_v1(
        p_operation_id, NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'installation activation event evidence is unavailable';
    END IF;
    IF v_operation.operation_kind = 'saas_customer_tenant_provision' THEN
        SELECT outcome.setup_artifact_id INTO v_artifact_id
        FROM platform.tenancy_customer_provisioning_outcomes AS outcome
        WHERE outcome.operation_id = p_operation_id;
        IF v_artifact_id IS NULL OR NOT
            platform.platform_async_installation_activation_evidence_assert_v1(
                p_operation_id, v_artifact_id
            )
        THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'initial admin delivery evidence is unavailable';
        END IF;
    END IF;
    IF NOT platform.tenancy_activation_step_consume_v1(
        p_operation_id, 'activation.complete'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant activation completion capability consume failed';
    END IF;
    PERFORM platform.tenancy_onboarding_complete_v1(
        p_operation_id, p_receipt_id, p_receipt_sha256,
        p_evidence_sha256, p_expected_revision
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_complete_v2(
    uuid,uuid,text,text,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_complete_v2(
    uuid,uuid,text,text,bigint
) TO forgecrud_runtime;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_activate_v2(
    p_operation_id uuid,
    p_lease_token_sha256 text,
    p_expected_revision bigint
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
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_claim_operation_id uuid;
    v_claim_tenant_id uuid;
    v_claim_operation_kind text;
    v_claim_step_plan_sha256 text;
    v_claim_migration_receipt_id uuid;
    v_claim_migration_receipt_sha256 text;
    v_claim_authority_ack_sha256 text;
    v_claim_revision bigint;
    v_claim_database_now timestamptz;
    v_claim_lease_until timestamptz;
    v_operation platform.tenancy_onboarding_operations%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant activation capability mint denied';
    END IF;
    SELECT claim.operation_id, claim.tenant_id, claim.operation_kind,
           claim.step_plan_sha256, claim.migration_receipt_id,
           claim.migration_receipt_sha256, claim.authority_ack_sha256,
           claim.revision, claim.database_now, claim.lease_until
    INTO STRICT v_claim_operation_id, v_claim_tenant_id,
        v_claim_operation_kind, v_claim_step_plan_sha256,
        v_claim_migration_receipt_id, v_claim_migration_receipt_sha256,
        v_claim_authority_ack_sha256, v_claim_revision,
        v_claim_database_now, v_claim_lease_until
    FROM platform.tenancy_onboarding_activate_v1(
        p_operation_id, p_lease_token_sha256, p_expected_revision
    ) AS claim;
    SELECT * INTO STRICT v_operation
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id = p_operation_id
    FOR UPDATE;
    IF v_operation.state IS DISTINCT FROM 'finalizing'
       OR v_operation.revision IS DISTINCT FROM v_claim_revision
       OR v_operation.tenant_id IS DISTINCT FROM v_claim_tenant_id
       OR v_operation.operation_kind IS DISTINCT FROM v_claim_operation_kind
       OR v_operation.step_plan_sha256
            IS DISTINCT FROM v_claim_step_plan_sha256
       OR NOT EXISTS (
           SELECT 1
           FROM platform.tenancy_onboarding_steps AS step
           WHERE step.operation_id = v_operation.operation_id
             AND step.command_key = 'billing.assign_tenant_base_offering'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant activation capability plan conflict';
    END IF;
    INSERT INTO platform.tenancy_activation_command_capabilities (
        backend_pid, transaction_id, operation_id, command_key,
        environment_id, tenant_id, operation_kind, deployment_mode,
        binding_purpose, manifest_sha256, step_plan_sha256,
        offering_version_id, offering_revision, offering_sha256,
        offering_effective_at, offering_source_sequence,
        offering_source_sha256, created_at
    ) VALUES (
        pg_catalog.pg_backend_pid(), pg_catalog.pg_current_xact_id(),
        v_operation.operation_id, 'billing.assign_tenant_base_offering',
        v_operation.environment_id, v_operation.tenant_id,
        v_operation.operation_kind, v_operation.deployment_mode,
        v_operation.binding_purpose, v_operation.manifest_sha256,
        v_operation.step_plan_sha256, v_operation.offering_version_id,
        v_operation.offering_revision, v_operation.offering_sha256,
        v_operation.offering_effective_at,
        v_operation.offering_source_sequence,
        v_operation.offering_source_sha256,
        pg_catalog.transaction_timestamp()
    );
    RETURN QUERY SELECT
        v_claim_operation_id, v_claim_tenant_id, v_claim_operation_kind,
        v_claim_step_plan_sha256, v_claim_migration_receipt_id,
        v_claim_migration_receipt_sha256, v_claim_authority_ack_sha256,
        v_claim_revision, v_claim_database_now, v_claim_lease_until;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_activate_v2(
    uuid,text,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v2(
    uuid,text,bigint
) TO harpy_runtime;

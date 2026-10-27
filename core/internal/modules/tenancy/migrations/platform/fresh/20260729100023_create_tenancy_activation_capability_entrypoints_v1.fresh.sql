-- +goose Up
-- Migration stream version: 20260729100023.
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_onboarding_activate_v3(
    p_operation_id uuid,
    p_lease_token_sha256 text,
    p_expected_revision bigint
)
RETURNS TABLE (
    operation_id uuid, tenant_id uuid, operation_kind text,
    step_plan_sha256 text, migration_receipt_id uuid,
    migration_receipt_sha256 text, authority_ack_sha256 text,
    revision bigint, database_now timestamptz, lease_until timestamptz
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
    SELECT claim.operation_id, claim.tenant_id, claim.operation_kind,
        claim.step_plan_sha256, claim.migration_receipt_id,
        claim.migration_receipt_sha256, claim.authority_ack_sha256,
        claim.revision, claim.database_now, claim.lease_until
    INTO STRICT v_claim_operation_id, v_claim_tenant_id,
        v_claim_operation_kind, v_claim_step_plan_sha256,
        v_claim_migration_receipt_id, v_claim_migration_receipt_sha256,
        v_claim_authority_ack_sha256, v_claim_revision,
        v_claim_database_now, v_claim_lease_until
    FROM platform.tenancy_onboarding_activate_v2(
        p_operation_id, p_lease_token_sha256, p_expected_revision
    ) AS claim;
    SELECT * INTO STRICT v_operation
    FROM platform.tenancy_onboarding_operations AS candidate
    WHERE candidate.operation_id = p_operation_id
    FOR UPDATE;
    IF v_operation.state IS DISTINCT FROM 'finalizing'
       OR v_operation.revision IS DISTINCT FROM v_claim_revision
       OR v_operation.step_plan_sha256
            IS DISTINCT FROM v_claim_step_plan_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant activation step plan conflict';
    END IF;
    IF EXISTS (
        SELECT required.command_key
        FROM (
            VALUES
                ('tenancy.activate_tenant', false),
                ('iam.initialize_tenant_identity', false),
                ('billing.assign_tenant_base_offering', false),
                ('billing.open_tenant_usage', false),
                ('systemresources.initialize_tenant_baseline', false),
                ('platformops.initialize_recovery_signing_identity', true),
                ('platformauthority.bind_owner', true),
                ('platformauthority.pin_protected_authority_head', false),
                ('activation.complete', false)
        ) AS required(command_key, environment_only)
        WHERE (NOT required.environment_only
               OR v_operation.operation_kind = 'environment_initialize')
          AND NOT EXISTS (
              SELECT 1
              FROM platform.tenancy_onboarding_steps AS persisted
              WHERE persisted.operation_id = v_operation.operation_id
                AND persisted.command_key = required.command_key
          )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant activation command manifest conflict';
    END IF;
    INSERT INTO platform.tenancy_activation_step_capabilities (
        backend_pid, transaction_id, operation_id, command_key,
        environment_id, tenant_id, schema_name, operation_kind,
        deployment_mode, binding_purpose, manifest_sha256,
        step_plan_sha256, created_at
    )
    SELECT pg_catalog.pg_backend_pid(), pg_catalog.pg_current_xact_id(),
        v_operation.operation_id, command.command_key,
        v_operation.environment_id, v_operation.tenant_id,
        v_operation.schema_name, v_operation.operation_kind,
        v_operation.deployment_mode, v_operation.binding_purpose,
        v_operation.manifest_sha256, v_operation.step_plan_sha256,
        pg_catalog.transaction_timestamp()
    FROM (
        VALUES
            ('tenancy.activate_tenant', false),
            ('iam.initialize_tenant_identity', false),
            ('billing.open_tenant_usage', false),
            ('systemresources.initialize_tenant_baseline', false),
            ('platformops.initialize_recovery_signing_identity', true),
            ('platformauthority.bind_owner', true),
            ('platformauthority.pin_protected_authority_head', false),
            ('activation.complete', false)
    ) AS command(command_key, environment_only)
    WHERE NOT command.environment_only
       OR v_operation.operation_kind = 'environment_initialize';
    RETURN QUERY SELECT v_claim_operation_id, v_claim_tenant_id,
        v_claim_operation_kind, v_claim_step_plan_sha256,
        v_claim_migration_receipt_id, v_claim_migration_receipt_sha256,
        v_claim_authority_ack_sha256, v_claim_revision,
        v_claim_database_now, v_claim_lease_until;
END;
$function$;

CREATE FUNCTION platform.tenancy_activation_initialize_v5(
    p_operation_id uuid, p_environment_id uuid, p_tenant_id uuid,
    p_tenant_name text, p_tenant_slug text, p_owner_user_id uuid,
    p_schema_name text, p_binding_purpose text, p_manifest_sha256 text,
    p_operation_kind text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_replayed boolean;
BEGIN
    IF NOT platform.tenancy_activation_step_assert_v1(
        p_operation_id, 'tenancy.activate_tenant'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant activation step capability required';
    END IF;
    v_replayed := platform.tenancy_activation_initialize_v4(
        p_operation_id, p_environment_id, p_tenant_id, p_tenant_name,
        p_tenant_slug, p_owner_user_id, p_schema_name, p_binding_purpose,
        p_manifest_sha256, p_operation_kind
    );
    IF NOT platform.tenancy_activation_step_consume_v1(
        p_operation_id, 'tenancy.activate_tenant'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant activation step capability consume failed';
    END IF;
    RETURN v_replayed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_onboarding_activate_v3(
    uuid,text,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_onboarding_activate_v3(
    uuid,text,bigint
) TO harpy_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v5(
    uuid,uuid,uuid,text,text,uuid,text,text,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v5(
    uuid,uuid,uuid,text,text,uuid,text,text,text,text
) TO harpy_runtime;

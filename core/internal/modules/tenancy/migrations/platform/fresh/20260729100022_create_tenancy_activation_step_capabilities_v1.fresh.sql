-- +goose Up
-- Migration stream version: 20260729100022.
CREATE TABLE platform.tenancy_activation_step_capabilities (
    backend_pid integer NOT NULL,
    transaction_id xid8 NOT NULL,
    operation_id uuid NOT NULL,
    command_key text NOT NULL,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    schema_name name NOT NULL,
    operation_kind text NOT NULL,
    deployment_mode text NOT NULL,
    binding_purpose text NOT NULL,
    manifest_sha256 text NOT NULL,
    step_plan_sha256 text NOT NULL,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (backend_pid, transaction_id, operation_id, command_key),
    FOREIGN KEY (operation_id)
        REFERENCES platform.tenancy_onboarding_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (backend_pid > 0),
    CHECK (command_key IN (
        'tenancy.activate_tenant',
        'iam.initialize_tenant_identity',
        'billing.open_tenant_usage',
        'systemresources.initialize_tenant_baseline',
        'platformops.initialize_recovery_signing_identity',
        'platformauthority.bind_owner',
        'platformauthority.pin_protected_authority_head',
        'activation.complete'
    )),
    CHECK (operation_kind IN (
        'environment_initialize', 'saas_customer_tenant_provision'
    )),
    CHECK (deployment_mode IN ('onpremise', 'saas')),
    CHECK (binding_purpose IN (
        'onpremise_primary', 'saas_control', 'saas_customer'
    )),
    CHECK (schema_name::text ~ '^t_[a-z2-7]{26}$'),
    CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    CHECK (step_plan_sha256 ~ '^[0-9a-f]{64}$')
);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_step_assert_v1(
    p_operation_id uuid,
    p_command_key text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN SESSION_USER = 'harpy_v2_runtime'
       AND CURRENT_USER = 'harpy_owner'
       AND EXISTS (
            SELECT 1
            FROM platform.tenancy_activation_step_capabilities AS capability
            JOIN platform.tenancy_onboarding_operations AS operation
              ON operation.operation_id = capability.operation_id
            JOIN platform.tenancy_onboarding_leases AS lease
              ON lease.operation_id = operation.operation_id
            WHERE capability.backend_pid = pg_catalog.pg_backend_pid()
              AND capability.transaction_id = pg_catalog.pg_current_xact_id()
              AND capability.operation_id = p_operation_id
              AND capability.command_key = p_command_key
              AND operation.state = 'finalizing'
              AND operation.environment_id = capability.environment_id
              AND operation.tenant_id = capability.tenant_id
              AND operation.schema_name = capability.schema_name
              AND operation.operation_kind = capability.operation_kind
              AND operation.deployment_mode = capability.deployment_mode
              AND operation.binding_purpose = capability.binding_purpose
              AND operation.manifest_sha256 = capability.manifest_sha256
              AND operation.step_plan_sha256 = capability.step_plan_sha256
              AND lease.lease_until > pg_catalog.statement_timestamp()
       );
END;
$function$;

CREATE FUNCTION platform.tenancy_activation_step_consume_v1(
    p_operation_id uuid,
    p_command_key text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_consumed boolean;
BEGIN
    IF NOT platform.tenancy_activation_step_assert_v1(
        p_operation_id, p_command_key
    ) THEN
        RETURN false;
    END IF;
    DELETE FROM platform.tenancy_activation_step_capabilities AS capability
    WHERE capability.backend_pid = pg_catalog.pg_backend_pid()
      AND capability.transaction_id = pg_catalog.pg_current_xact_id()
      AND capability.operation_id = p_operation_id
      AND capability.command_key = p_command_key
    RETURNING true INTO v_consumed;
    RETURN COALESCE(v_consumed, false);
END;
$function$;

CREATE FUNCTION platform.tenancy_activation_step_context_assert_v1(
    p_operation_id uuid,
    p_command_key text,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_schema_name name,
    p_operation_kind text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    RETURN platform.tenancy_activation_step_assert_v1(
        p_operation_id, p_command_key
    ) AND EXISTS (
        SELECT 1
        FROM platform.tenancy_activation_step_capabilities AS capability
        WHERE capability.backend_pid = pg_catalog.pg_backend_pid()
          AND capability.transaction_id = pg_catalog.pg_current_xact_id()
          AND capability.operation_id = p_operation_id
          AND capability.command_key = p_command_key
          AND capability.environment_id = p_environment_id
          AND capability.tenant_id = p_tenant_id
          AND capability.schema_name = p_schema_name
          AND capability.operation_kind = p_operation_kind
    );
END;
$function$;

CREATE FUNCTION platform.tenancy_activation_step_commit_guard_v1()
RETURNS trigger
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM platform.tenancy_activation_step_capabilities AS capability
        WHERE capability.backend_pid = NEW.backend_pid
          AND capability.transaction_id = NEW.transaction_id
          AND capability.operation_id = NEW.operation_id
          AND capability.command_key = NEW.command_key
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'tenancy activation step capability was not consumed';
    END IF;
    RETURN NULL;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_activation_step_assert_v1(
    uuid,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_step_assert_v1(
    uuid,text
) FROM harpy_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_activation_step_consume_v1(
    uuid,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_step_consume_v1(
    uuid,text
) FROM harpy_runtime;
REVOKE ALL ON FUNCTION platform.tenancy_activation_step_context_assert_v1(
    uuid,text,uuid,uuid,name,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION
    platform.tenancy_activation_step_context_assert_v1(
        uuid,text,uuid,uuid,name,text
    ) FROM harpy_runtime;
REVOKE ALL ON FUNCTION
    platform.tenancy_activation_step_commit_guard_v1() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION
    platform.tenancy_activation_step_commit_guard_v1()
    FROM harpy_runtime;

CREATE CONSTRAINT TRIGGER tenancy_activation_step_commit_guard
AFTER INSERT ON platform.tenancy_activation_step_capabilities
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION
    platform.tenancy_activation_step_commit_guard_v1();
ALTER TABLE platform.tenancy_activation_step_capabilities
    ENABLE ALWAYS TRIGGER tenancy_activation_step_commit_guard;

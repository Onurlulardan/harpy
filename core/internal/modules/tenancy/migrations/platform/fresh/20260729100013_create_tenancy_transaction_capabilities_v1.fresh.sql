-- +goose Up
CREATE TABLE platform.tenancy_activation_command_capabilities (
    backend_pid integer NOT NULL,
    transaction_id xid8 NOT NULL,
    operation_id uuid NOT NULL,
    command_key text NOT NULL,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    operation_kind text NOT NULL,
    deployment_mode text NOT NULL,
    binding_purpose text NOT NULL,
    manifest_sha256 text NOT NULL,
    step_plan_sha256 text NOT NULL,
    offering_version_id uuid NOT NULL,
    offering_revision bigint NOT NULL,
    offering_sha256 text NOT NULL,
    offering_effective_at timestamptz NOT NULL,
    offering_source_sequence bigint,
    offering_source_sha256 text,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (backend_pid, transaction_id, operation_id, command_key),
    FOREIGN KEY (operation_id)
        REFERENCES platform.tenancy_onboarding_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (backend_pid > 0),
    CHECK (command_key = 'billing.assign_tenant_base_offering'),
    CHECK (operation_kind IN (
        'environment_initialize', 'saas_customer_tenant_provision'
    )),
    CHECK (deployment_mode IN ('onpremise', 'saas')),
    CHECK (binding_purpose IN (
        'onpremise_primary', 'saas_control', 'saas_customer'
    )),
    CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    CHECK (step_plan_sha256 ~ '^[0-9a-f]{64}$'),
    CHECK (offering_revision > 0),
    CHECK (offering_sha256 ~ '^[0-9a-f]{64}$'),
    CHECK (
        (deployment_mode = 'onpremise'
         AND offering_source_sequence > 0
         AND offering_source_sha256 ~ '^[0-9a-f]{64}$')
        OR (deployment_mode = 'saas'
            AND offering_source_sequence IS NULL
            AND offering_source_sha256 IS NULL)
    )
);

CREATE TABLE platform.tenancy_authority_reconcile_scopes (
    backend_pid integer NOT NULL,
    transaction_id xid8 NOT NULL,
    operation_id uuid NOT NULL,
    environment_id uuid NOT NULL,
    manifest_sha256 text NOT NULL,
    step_plan_sha256 text NOT NULL,
    owner_key text NOT NULL,
    created_at timestamptz NOT NULL,
    PRIMARY KEY (backend_pid, transaction_id, operation_id, owner_key),
    FOREIGN KEY (operation_id)
        REFERENCES platform.tenancy_onboarding_operations (operation_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (backend_pid > 0),
    CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    CHECK (step_plan_sha256 ~ '^[0-9a-f]{64}$'),
    CHECK (owner_key IN ('billing', 'analytics'))
);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_command_consume_v1(
    p_operation_id uuid, p_command_key text, p_environment_id uuid,
    p_tenant_id uuid, p_operation_kind text, p_deployment_mode text,
    p_binding_purpose text, p_manifest_sha256 text,
    p_offering_version_id uuid, p_offering_revision bigint,
    p_offering_sha256 text, p_offering_effective_at timestamptz,
    p_offering_source_sequence bigint, p_offering_source_sha256 text
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog AS $function$
DECLARE
    v_consumed boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RETURN false;
    END IF;
    DELETE FROM platform.tenancy_activation_command_capabilities AS capability
    WHERE capability.backend_pid = pg_catalog.pg_backend_pid()
      AND capability.transaction_id = pg_catalog.pg_current_xact_id()
      AND capability.operation_id = p_operation_id
      AND capability.command_key = p_command_key
      AND capability.environment_id = p_environment_id
      AND capability.tenant_id = p_tenant_id
      AND capability.operation_kind = p_operation_kind
      AND capability.deployment_mode = p_deployment_mode
      AND capability.binding_purpose = p_binding_purpose
      AND capability.manifest_sha256 = p_manifest_sha256
      AND capability.offering_version_id = p_offering_version_id
      AND capability.offering_revision = p_offering_revision
      AND capability.offering_sha256 = p_offering_sha256
      AND capability.offering_effective_at = p_offering_effective_at
      AND capability.offering_source_sequence
            IS NOT DISTINCT FROM p_offering_source_sequence
      AND capability.offering_source_sha256
            IS NOT DISTINCT FROM p_offering_source_sha256
    RETURNING true INTO v_consumed;
    RETURN coalesce(v_consumed, false);
END;
$function$;

CREATE FUNCTION platform.tenancy_authority_reconcile_scope_assert_v1(
    p_owner_key text
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog AS $function$
BEGIN
    RETURN SESSION_USER = 'harpy_v2_runtime'
       AND CURRENT_USER = 'harpy_owner'
       AND p_owner_key IN ('billing', 'analytics')
       AND EXISTS (
            SELECT 1
            FROM platform.tenancy_authority_reconcile_scopes AS scope
            WHERE scope.backend_pid = pg_catalog.pg_backend_pid()
              AND scope.transaction_id = pg_catalog.pg_current_xact_id()
              AND scope.owner_key = p_owner_key
       );
END;
$function$;

CREATE FUNCTION platform.tenancy_activation_capability_commit_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog AS $function$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM platform.tenancy_activation_command_capabilities AS capability
        WHERE capability.backend_pid = NEW.backend_pid
          AND capability.transaction_id = NEW.transaction_id
          AND capability.operation_id = NEW.operation_id
          AND capability.command_key = NEW.command_key
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'tenancy activation command capability was not consumed';
    END IF;
    RETURN NULL;
END;
$function$;

CREATE FUNCTION platform.tenancy_authority_scope_commit_guard_v1()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog AS $function$
BEGIN
    IF EXISTS (
        SELECT 1 FROM platform.tenancy_authority_reconcile_scopes AS scope
        WHERE scope.backend_pid = NEW.backend_pid
          AND scope.transaction_id = NEW.transaction_id
          AND scope.operation_id = NEW.operation_id
          AND scope.owner_key = NEW.owner_key
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'tenancy authority reconcile scope was not cleared';
    END IF;
    RETURN NULL;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_activation_command_consume_v1(
    uuid,text,uuid,uuid,text,text,text,text,uuid,bigint,text,timestamptz,bigint,text
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_activation_command_consume_v1(
    uuid,text,uuid,uuid,text,text,text,text,uuid,bigint,text,timestamptz,bigint,text
) FROM harpy_runtime;
REVOKE ALL ON FUNCTION
    platform.tenancy_authority_reconcile_scope_assert_v1(text)
    FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION
    platform.tenancy_authority_reconcile_scope_assert_v1(text)
    FROM harpy_runtime;
REVOKE ALL ON FUNCTION
    platform.tenancy_activation_capability_commit_guard_v1() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION
    platform.tenancy_activation_capability_commit_guard_v1()
    FROM harpy_runtime;
REVOKE ALL ON FUNCTION
    platform.tenancy_authority_scope_commit_guard_v1() FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION platform.tenancy_authority_scope_commit_guard_v1()
    FROM harpy_runtime;

CREATE CONSTRAINT TRIGGER tenancy_activation_capability_commit_guard
AFTER INSERT ON platform.tenancy_activation_command_capabilities
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION
    platform.tenancy_activation_capability_commit_guard_v1();
ALTER TABLE platform.tenancy_activation_command_capabilities
    ENABLE ALWAYS TRIGGER tenancy_activation_capability_commit_guard;

CREATE CONSTRAINT TRIGGER tenancy_authority_scope_commit_guard
AFTER INSERT ON platform.tenancy_authority_reconcile_scopes
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_authority_scope_commit_guard_v1();
ALTER TABLE platform.tenancy_authority_reconcile_scopes
    ENABLE ALWAYS TRIGGER tenancy_authority_scope_commit_guard;

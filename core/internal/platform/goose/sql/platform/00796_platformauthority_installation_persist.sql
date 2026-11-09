-- +goose Up
CREATE TABLE IF NOT EXISTS platform.platformauthority_installation_owner_bindings (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    schema_name name NOT NULL,
    deployment_mode text NOT NULL,
    instance_owner_binding_id uuid NOT NULL,
    protected_admin_binding_id uuid NOT NULL,
    registry_binding_id uuid NOT NULL,
    manifest_sha256 text NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE TABLE IF NOT EXISTS platform.platformauthority_installation_head_pins (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    schema_name name NOT NULL,
    registry_binding_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    generation bigint NOT NULL,
    previous_head_sha256 text NOT NULL,
    state_sha256 text NOT NULL,
    proposal_sha256 text NOT NULL,
    ack_sha256 text NOT NULL,
    manifest_sha256 text NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT platformauthority_installation_head_pins_generation_check CHECK (generation > 0)
);

REVOKE ALL ON TABLE platform.platformauthority_installation_owner_bindings FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON TABLE platform.platformauthority_installation_head_pins FROM PUBLIC, harpy_api, harpy_worker;

DROP FUNCTION platform.platformauthority_installation_owner_bind_v3(uuid, uuid, uuid, name, text, uuid, uuid, uuid, text);
DROP FUNCTION platform.platformauthority_installation_head_pin_v3(uuid, uuid, uuid, name, uuid, text, uuid, bigint, text, text, text, text, text);

-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_installation_owner_bind_v3(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name, in_deployment_mode text,
    in_instance_owner_binding_id uuid, in_protected_admin_binding_id uuid, in_registry_binding_id uuid, in_manifest_sha256 text
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE;
BEGIN
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id;
    IF NOT FOUND OR operation.tenant_id <> in_tenant_id OR operation.environment_id IS DISTINCT FROM in_environment_id
        OR operation.schema_name IS DISTINCT FROM in_schema_name::text OR operation.deployment_mode IS DISTINCT FROM in_deployment_mode
        OR operation.instance_owner_binding_id IS DISTINCT FROM in_instance_owner_binding_id
        OR operation.protected_admin_binding_id IS DISTINCT FROM in_protected_admin_binding_id
        OR operation.registry_binding_id IS DISTINCT FROM in_registry_binding_id OR operation.manifest_sha256 IS DISTINCT FROM in_manifest_sha256
        OR NOT platform.tenancy_activation_assert_v2(in_operation_id, in_tenant_id, in_schema_name) THEN
        RAISE EXCEPTION 'platform authority owner binding does not match the activation' USING ERRCODE = '42501';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.platformauthority_installation_owner_bindings WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id
        AND instance_owner_binding_id = in_instance_owner_binding_id AND manifest_sha256 = in_manifest_sha256) THEN
        RETURN true;
    END IF;
    INSERT INTO platform.platformauthority_installation_owner_bindings (
        operation_id, environment_id, tenant_id, schema_name, deployment_mode, instance_owner_binding_id,
        protected_admin_binding_id, registry_binding_id, manifest_sha256, created_at
    ) VALUES (
        in_operation_id, in_environment_id, in_tenant_id, in_schema_name, in_deployment_mode, in_instance_owner_binding_id,
        in_protected_admin_binding_id, in_registry_binding_id, in_manifest_sha256, pg_catalog.clock_timestamp()
    );
    PERFORM platform.tenancy_activation_step_consume_v1(in_operation_id, 'platformauthority.bind_owner', in_tenant_id, in_schema_name);
    RETURN false;
END;
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_installation_head_pin_v3(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name, in_registry_binding_id uuid,
    in_scope_kind text, in_scope_id uuid, in_generation bigint, in_previous_head_sha256 text, in_state_sha256 text,
    in_proposal_sha256 text, in_ack_sha256 text, in_manifest_sha256 text
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE;
BEGIN
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id;
    IF NOT FOUND OR operation.deployment_mode = 'onpremise' OR operation.tenant_id <> in_tenant_id
        OR operation.environment_id IS DISTINCT FROM in_environment_id OR operation.schema_name IS DISTINCT FROM in_schema_name::text
        OR operation.registry_binding_id IS DISTINCT FROM in_registry_binding_id OR operation.authority_scope_kind IS DISTINCT FROM in_scope_kind
        OR operation.authority_scope_id IS DISTINCT FROM in_scope_id OR operation.authority_generation IS DISTINCT FROM in_generation
        OR operation.manifest_sha256 IS DISTINCT FROM in_manifest_sha256 OR in_generation < 1
        OR NOT platform.tenancy_activation_assert_v2(in_operation_id, in_tenant_id, in_schema_name) THEN
        RAISE EXCEPTION 'platform authority head pin does not match the activation' USING ERRCODE = '42501';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.platformauthority_installation_head_pins WHERE operation_id = in_operation_id AND generation = in_generation AND ack_sha256 = in_ack_sha256) THEN
        RETURN true;
    END IF;
    INSERT INTO platform.platformauthority_installation_head_pins (
        operation_id, environment_id, tenant_id, schema_name, registry_binding_id, scope_kind, scope_id, generation,
        previous_head_sha256, state_sha256, proposal_sha256, ack_sha256, manifest_sha256, created_at
    ) VALUES (
        in_operation_id, in_environment_id, in_tenant_id, in_schema_name, in_registry_binding_id, in_scope_kind, in_scope_id, in_generation,
        COALESCE(in_previous_head_sha256, ''), in_state_sha256, in_proposal_sha256, in_ack_sha256, in_manifest_sha256, pg_catalog.clock_timestamp()
    );
    PERFORM platform.tenancy_activation_step_consume_v1(in_operation_id, 'platformauthority.pin_protected_authority_head', in_tenant_id, in_schema_name);
    RETURN false;
END;
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_installation_owner_bind_v3(uuid, uuid, uuid, name, text, uuid, uuid, uuid, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_installation_head_pin_v3(uuid, uuid, uuid, name, uuid, text, uuid, bigint, text, text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformauthority_installation_owner_bind_v3(uuid, uuid, uuid, name, text, uuid, uuid, uuid, text) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.platformauthority_installation_head_pin_v3(uuid, uuid, uuid, name, uuid, text, uuid, bigint, text, text, text, text, text) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platformauthority_installation_head_pin_v3(uuid, uuid, uuid, name, uuid, text, uuid, bigint, text, text, text, text, text);
DROP FUNCTION platform.platformauthority_installation_owner_bind_v3(uuid, uuid, uuid, name, text, uuid, uuid, uuid, text);
DROP TABLE IF EXISTS platform.platformauthority_installation_head_pins;
DROP TABLE IF EXISTS platform.platformauthority_installation_owner_bindings;
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_installation_owner_bind_v3(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name, in_deployment_mode text, in_instance_owner_binding_id uuid, in_protected_admin_binding_id uuid, in_registry_binding_id uuid, in_manifest_sha256 text)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_installation_head_pin_v3(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_schema_name name, in_registry_binding_id uuid, in_scope_kind text, in_scope_id uuid, in_generation bigint, in_previous_head_sha256 text, in_state_sha256 text, in_proposal_sha256 text, in_ack_sha256 text, in_manifest_sha256 text)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platformauthority_installation_owner_bind_v3(uuid, uuid, uuid, name, text, uuid, uuid, uuid, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.platformauthority_installation_head_pin_v3(uuid, uuid, uuid, name, uuid, text, uuid, bigint, text, text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformauthority_installation_owner_bind_v3(uuid, uuid, uuid, name, text, uuid, uuid, uuid, text) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.platformauthority_installation_head_pin_v3(uuid, uuid, uuid, name, uuid, text, uuid, bigint, text, text, text, text, text) TO harpy_api, harpy_worker;

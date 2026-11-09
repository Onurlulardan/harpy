-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_digest_hex_v1(in_payload text)
RETURNS text LANGUAGE sql IMMUTABLE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$ SELECT pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(in_payload, 'UTF8')), 'hex'); $fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_digest_uuid_v1(in_payload text)
RETURNS uuid LANGUAGE sql IMMUTABLE SECURITY INVOKER SET search_path TO pg_catalog
AS $fn$
    SELECT pg_catalog.encode(pg_catalog.set_byte(pg_catalog.set_byte(raw.value, 6,
        (pg_catalog.get_byte(raw.value, 6) & 15) | 112), 8,
        (pg_catalog.get_byte(raw.value, 8) & 63) | 128), 'hex')::uuid
    FROM (SELECT pg_catalog.substr(pg_catalog.sha256(pg_catalog.convert_to(in_payload, 'UTF8')), 1, 16) AS value) AS raw;
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_step_ordinal_v1(in_operation_id uuid, in_command_name text, in_tenant_id uuid, in_schema_name name)
RETURNS integer LANGUAGE sql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
    SELECT capability.ordinal
    FROM platform.tenancy_activation_step_capabilities AS capability
    JOIN platform.tenancy_onboarding_operations AS operation
      ON operation.operation_id = capability.operation_id
     AND operation.tenant_id = capability.tenant_id
     AND operation.schema_name = capability.schema_name
     AND operation.kind = capability.profile_kind
     AND operation.manifest_sha256 = capability.manifest_sha256
     AND operation.environment_id = capability.environment_id
    WHERE capability.operation_id = in_operation_id AND capability.command_name = in_command_name
      AND capability.tenant_id = in_tenant_id AND capability.schema_name = in_schema_name::text
      AND capability.session_user_name = session_user
      AND capability.backend_pid = pg_catalog.pg_backend_pid()
      AND capability.transaction_id IS NOT DISTINCT FROM pg_catalog.pg_current_xact_id();
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_step_proof_v1(in_operation_id uuid, in_command_name text, in_tenant_id uuid, in_schema_name name)
RETURNS boolean LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE proved boolean;
BEGIN
    IF in_schema_name::text !~ '^tenant_[0-9a-f]{32}$' THEN RETURN false; END IF;
    IF in_command_name = 'tenancy.activate_tenant' THEN
        SELECT true INTO proved FROM platform.tenancy_activation_facts
        WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id AND schema_name = in_schema_name::text;
    ELSIF in_command_name = 'iam.initialize_tenant_identity' THEN
        EXECUTE pg_catalog.format('SELECT true FROM %I.iam_installation_credentials WHERE operation_id = $1 AND tenant_id = $2', in_schema_name) INTO proved USING in_operation_id, in_tenant_id;
    ELSIF in_command_name = 'billing.assign_tenant_base_offering' THEN
        SELECT true INTO proved FROM platform.billing_installation_base_assignments WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id;
    ELSIF in_command_name = 'billing.open_tenant_usage' THEN
        SELECT true INTO proved FROM platform.billing_installation_usage_openings WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id;
    ELSIF in_command_name = 'systemresources.initialize_tenant_baseline' THEN
        EXECUTE pg_catalog.format('SELECT true FROM %I.system_resource_layout_installation_baselines WHERE operation_id = $1 AND tenant_id = $2', in_schema_name) INTO proved USING in_operation_id, in_tenant_id;
    ELSIF in_command_name = 'platformops.initialize_recovery_signing_identity' THEN
        SELECT true INTO proved FROM platform.platformops_installation_recovery_identities WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id;
    ELSIF in_command_name = 'platformauthority.bind_owner' THEN
        SELECT true INTO proved FROM platform.platformauthority_installation_owner_bindings WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id;
    ELSIF in_command_name = 'platformauthority.pin_protected_authority_head' THEN
        SELECT true INTO proved FROM platform.platformauthority_installation_head_pins WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id;
    ELSIF in_command_name = 'activation.complete' THEN
        SELECT true INTO proved FROM platform.tenancy_onboarding_receipts
        WHERE operation_id = in_operation_id AND receipt_kind = 'activation_completed';
    END IF;
    RETURN COALESCE(proved, false);
END;
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_step_context_assert_v1(in_operation_id uuid, in_command_name text, in_tenant_id uuid, in_schema_name name)
RETURNS boolean LANGUAGE sql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$ SELECT platform.tenancy_activation_step_ordinal_v1(in_operation_id, in_command_name, in_tenant_id, in_schema_name) IS NOT NULL; $fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_step_consume_v1(in_operation_id uuid, in_command_name text, in_tenant_id uuid, in_schema_name name)
RETURNS void LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE step_ordinal integer;
BEGIN
    step_ordinal := platform.tenancy_activation_step_ordinal_v1(in_operation_id, in_command_name, in_tenant_id, in_schema_name);
    IF step_ordinal IS NULL OR NOT platform.tenancy_activation_step_proof_v1(in_operation_id, in_command_name, in_tenant_id, in_schema_name) THEN
        RAISE EXCEPTION 'activation step capability is not valid for this transaction' USING ERRCODE = '42501';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.tenancy_activation_step_capabilities WHERE operation_id = in_operation_id AND ordinal < step_ordinal) THEN
        RAISE EXCEPTION 'activation step is out of order' USING ERRCODE = '23514';
    END IF;
    DELETE FROM platform.tenancy_activation_step_capabilities WHERE operation_id = in_operation_id AND command_name = in_command_name;
END;
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_step_commit_guard_v1()
RETURNS trigger LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
BEGIN
    IF EXISTS (SELECT 1 FROM platform.tenancy_activation_step_capabilities
        WHERE operation_id = NEW.operation_id AND backend_pid = pg_catalog.pg_backend_pid()
          AND transaction_id IS NOT DISTINCT FROM pg_catalog.pg_current_xact_id()) THEN
        RAISE EXCEPTION 'activation step capability was not consumed' USING ERRCODE = '23514';
    END IF;
    RETURN NULL;
END;
$fn$;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS tenancy_activation_step_capabilities_commit ON platform.tenancy_activation_step_capabilities;
CREATE CONSTRAINT TRIGGER tenancy_activation_step_capabilities_commit
    AFTER INSERT ON platform.tenancy_activation_step_capabilities
    DEFERRABLE INITIALLY DEFERRED FOR EACH ROW
    EXECUTE FUNCTION platform.tenancy_activation_step_commit_guard_v1();

REVOKE ALL ON FUNCTION platform.tenancy_activation_digest_hex_v1(text) FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON FUNCTION platform.tenancy_activation_digest_uuid_v1(text) FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON FUNCTION platform.tenancy_activation_step_ordinal_v1(uuid, text, uuid, name) FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON FUNCTION platform.tenancy_activation_step_proof_v1(uuid, text, uuid, name) FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON FUNCTION platform.tenancy_activation_step_commit_guard_v1() FROM PUBLIC, harpy_api, harpy_worker;
REVOKE ALL ON FUNCTION platform.tenancy_activation_step_context_assert_v1(uuid, text, uuid, name) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_activation_step_consume_v1(uuid, text, uuid, name) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_step_context_assert_v1(uuid, text, uuid, name) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_step_consume_v1(uuid, text, uuid, name) TO harpy_api, harpy_worker;

-- +goose Down
DROP TRIGGER IF EXISTS tenancy_activation_step_capabilities_commit ON platform.tenancy_activation_step_capabilities;
DROP FUNCTION IF EXISTS platform.tenancy_activation_step_commit_guard_v1();
DROP FUNCTION IF EXISTS platform.tenancy_activation_step_consume_v1(uuid, text, uuid, name);
DROP FUNCTION IF EXISTS platform.tenancy_activation_step_context_assert_v1(uuid, text, uuid, name);
DROP FUNCTION IF EXISTS platform.tenancy_activation_step_proof_v1(uuid, text, uuid, name);
DROP FUNCTION IF EXISTS platform.tenancy_activation_step_ordinal_v1(uuid, text, uuid, name);
DROP FUNCTION IF EXISTS platform.tenancy_activation_digest_uuid_v1(text);
DROP FUNCTION IF EXISTS platform.tenancy_activation_digest_hex_v1(text);

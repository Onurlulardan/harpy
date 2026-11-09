-- +goose Up
DROP FUNCTION platform.tenancy_activation_initialize_v5(uuid, uuid, uuid, text, text, uuid, text, text, text, text);
DROP FUNCTION platform.tenancy_activation_assert_v2(uuid, uuid, name);

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_initialize_v5(
    in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_name text, in_slug text,
    in_owner_user_id uuid, in_schema_name text, in_binding_purpose text, in_manifest_sha256 text, in_kind text
) RETURNS boolean LANGUAGE plpgsql VOLATILE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
DECLARE operation platform.tenancy_onboarding_operations%ROWTYPE; tenant platform.tenants%ROWTYPE;
BEGIN
    SELECT * INTO operation FROM platform.tenancy_onboarding_operations WHERE operation_id = in_operation_id;
    IF NOT FOUND OR operation.tenant_id <> in_tenant_id OR operation.environment_id IS DISTINCT FROM in_environment_id
        OR operation.schema_name IS DISTINCT FROM in_schema_name OR operation.kind IS DISTINCT FROM in_kind
        OR operation.binding_purpose IS DISTINCT FROM in_binding_purpose OR operation.manifest_sha256 IS DISTINCT FROM in_manifest_sha256
        OR operation.first_owner_user_id <> in_owner_user_id THEN
        RAISE EXCEPTION 'tenant activation does not match the onboarding operation' USING ERRCODE = '42501';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.tenancy_activation_facts WHERE operation_id = in_operation_id AND tenant_id = in_tenant_id
        AND schema_name = in_schema_name AND manifest_sha256 = in_manifest_sha256 AND owner_user_id = in_owner_user_id) THEN
        RETURN true;
    END IF;
    IF EXISTS (SELECT 1 FROM platform.tenancy_activation_facts WHERE operation_id = in_operation_id) THEN
        RAISE EXCEPTION 'tenant activation fact belongs to another tenant' USING ERRCODE = '42501';
    END IF;
    INSERT INTO platform.tenants (id, name, slug, status, schema_name, settings_revision, lifecycle_revision, created_at, updated_at)
    VALUES (in_tenant_id, in_name, in_slug, 'ACTIVE', in_schema_name, 1, 1, pg_catalog.clock_timestamp(), pg_catalog.clock_timestamp())
    ON CONFLICT (id) DO NOTHING;
    SELECT * INTO tenant FROM platform.tenants WHERE id = in_tenant_id;
    IF NOT FOUND OR tenant.schema_name <> in_schema_name OR tenant.slug <> in_slug OR tenant.deleted_at IS NOT NULL THEN
        RAISE EXCEPTION 'tenant row does not match the activation' USING ERRCODE = '23514';
    END IF;
    INSERT INTO platform.tenancy_activation_facts (
        operation_id, environment_id, tenant_id, owner_user_id, schema_name, kind, name, slug, binding_purpose, manifest_sha256, activated_at
    ) VALUES (
        in_operation_id, in_environment_id, in_tenant_id, in_owner_user_id, in_schema_name, in_kind, in_name, in_slug,
        in_binding_purpose, in_manifest_sha256, pg_catalog.clock_timestamp()
    );
    PERFORM platform.tenancy_activation_step_consume_v1(in_operation_id, 'tenancy.activate_tenant', in_tenant_id, in_schema_name::name);
    RETURN false;
END;
$fn$;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_assert_v2(p_operation_id uuid, p_tenant_id uuid, p_schema_name name)
RETURNS boolean LANGUAGE sql STABLE SECURITY DEFINER SET search_path TO pg_catalog
AS $fn$
    SELECT EXISTS (
        SELECT 1 FROM platform.tenancy_activation_facts AS fact
        JOIN platform.tenancy_onboarding_operations AS operation ON operation.operation_id = fact.operation_id
        JOIN platform.tenants AS tenant ON tenant.id = fact.tenant_id
        WHERE fact.operation_id = p_operation_id AND fact.tenant_id = p_tenant_id AND fact.schema_name = p_schema_name::text
          AND operation.tenant_id = p_tenant_id AND operation.schema_name = p_schema_name::text
          AND tenant.schema_name = p_schema_name::text AND tenant.deleted_at IS NULL AND tenant.status = 'ACTIVE'
    );
$fn$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v5(uuid, uuid, uuid, text, text, uuid, text, text, text, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_activation_assert_v2(uuid, uuid, name) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v5(uuid, uuid, uuid, text, text, uuid, text, text, text, text) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_assert_v2(uuid, uuid, name) TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.tenancy_activation_assert_v2(uuid, uuid, name);
DROP FUNCTION platform.tenancy_activation_initialize_v5(uuid, uuid, uuid, text, text, uuid, text, text, text, text);
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_initialize_v5(in_operation_id uuid, in_environment_id uuid, in_tenant_id uuid, in_name text, in_slug text, in_owner_user_id uuid, in_schema_name text, in_binding_purpose text, in_manifest_sha256 text, in_kind text)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_assert_v2(p_operation_id uuid, p_tenant_id uuid, p_schema_name name)
RETURNS boolean LANGUAGE sql VOLATILE SECURITY INVOKER SET search_path TO pg_catalog AS $fn$ SELECT false; $fn$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v5(uuid, uuid, uuid, text, text, uuid, text, text, text, text) FROM PUBLIC;
REVOKE ALL ON FUNCTION platform.tenancy_activation_assert_v2(uuid, uuid, name) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v5(uuid, uuid, uuid, text, text, uuid, text, text, text, text) TO harpy_api, harpy_worker;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_assert_v2(uuid, uuid, name) TO harpy_api, harpy_worker;

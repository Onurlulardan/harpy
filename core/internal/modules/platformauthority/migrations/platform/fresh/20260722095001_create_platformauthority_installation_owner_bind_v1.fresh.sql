-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformauthority_installation_owner_bind_v1(
    p_operation_id uuid,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_schema_name name,
    p_deployment_mode text,
    p_instance_owner_binding_id uuid,
    p_protected_admin_binding_id uuid,
    p_registry_binding_id uuid,
    p_manifest_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.platformauthority_installation_owner_bindings%ROWTYPE;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'harpy-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_environment_id IS NULL
       OR p_tenant_id IS NULL OR p_schema_name IS NULL
       OR p_instance_owner_binding_id IS NULL OR p_protected_admin_binding_id IS NULL
       OR p_registry_binding_id IS NULL
       OR p_deployment_mode NOT IN ('onpremise', 'saas')
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'installation owner binding denied';
    END IF;
    IF NOT platform.tenancy_activation_assert_v1(
        p_operation_id, p_tenant_id, p_schema_name
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'installation owner activation conflicts';
    END IF;
    PERFORM 1
    FROM platform.platformauthority_bindings AS binding
    WHERE binding.binding_id = p_registry_binding_id
      AND binding.environment_id = p_environment_id
      AND binding.deployment_mode = p_deployment_mode
      AND binding.signer_status = 'active'
    FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'protected authority binding is not active';
    END IF;
    SELECT * INTO v_existing
    FROM platform.platformauthority_installation_owner_bindings AS owner_binding
    WHERE owner_binding.operation_id = p_operation_id
    FOR SHARE;
    IF FOUND THEN
        IF ROW(v_existing.environment_id, v_existing.tenant_id,
            v_existing.instance_owner_binding_id,
            v_existing.protected_admin_binding_id,
            v_existing.registry_binding_id, v_existing.manifest_sha256)
           IS DISTINCT FROM ROW(p_environment_id, p_tenant_id,
            p_instance_owner_binding_id, p_protected_admin_binding_id,
            p_registry_binding_id, p_manifest_sha256) THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'installation owner binding replay conflicts';
        END IF;
        RETURN true;
    END IF;
    INSERT INTO platform.platformauthority_installation_owner_bindings (
        operation_id, environment_id, tenant_id, instance_owner_binding_id,
        protected_admin_binding_id, registry_binding_id,
        manifest_sha256, created_at
    ) VALUES (
        p_operation_id, p_environment_id, p_tenant_id,
        p_instance_owner_binding_id, p_protected_admin_binding_id,
        p_registry_binding_id, p_manifest_sha256,
        pg_catalog.statement_timestamp()
    );
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformauthority_installation_owner_bind_v1(
    uuid,uuid,uuid,name,text,uuid,uuid,uuid,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformauthority_installation_owner_bind_v1(
    uuid,uuid,uuid,name,text,uuid,uuid,uuid,text
) TO harpy_runtime;

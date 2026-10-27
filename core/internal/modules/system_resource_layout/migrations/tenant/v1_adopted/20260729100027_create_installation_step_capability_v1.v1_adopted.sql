-- +goose Up
-- Migration stream version: 20260729100027.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_installation_assert_v2(
    p_operation_id uuid,
    p_tenant_id uuid,
    p_schema_name name
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_environment_id uuid;
    v_operation_kind text;
BEGIN
    SELECT operation.environment_id, operation.operation_kind
    INTO v_environment_id, v_operation_kind
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id = p_operation_id;
    IF NOT FOUND OR NOT platform.tenancy_activation_step_context_assert_v1(
        p_operation_id, 'systemresources.initialize_tenant_baseline',
        v_environment_id, p_tenant_id, p_schema_name, v_operation_kind
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'system resource installation capability required';
    END IF;
    PERFORM 1
    FROM platform.tenancy_activation_facts AS fact
    JOIN platform.tenants AS tenant ON tenant.id = fact.tenant_id
    WHERE fact.operation_id = p_operation_id
      AND fact.tenant_id = p_tenant_id
      AND tenant.schema_name = p_schema_name
      AND tenant.status = 'ACTIVE'
      AND tenant.deleted_at IS NULL;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'system resource installation target conflicts';
    END IF;
    RETURN true;
END;
$function$;

CREATE FUNCTION system_resource_layout_installation_record_v2(
    p_operation_id uuid,
    p_tenant_id uuid,
    p_user_revision_id uuid,
    p_organization_revision_id uuid,
    p_role_revision_id uuid,
    p_now timestamptz,
    p_schema_name name
)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_digest bytea;
BEGIN
    PERFORM system_resource_layout_installation_assert_v2(
        p_operation_id, p_tenant_id, p_schema_name
    );
    v_digest := system_resource_layout_installation_record_v1(
        p_operation_id, p_tenant_id, p_user_revision_id,
        p_organization_revision_id, p_role_revision_id, p_now
    );
    IF NOT platform.tenancy_activation_step_consume_v1(
        p_operation_id, 'systemresources.initialize_tenant_baseline'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'system resource installation capability consume failed';
    END IF;
    RETURN v_digest;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_installation_assert_v2(
    uuid,uuid,name
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_installation_assert_v2(
    uuid,uuid,name
) TO harpy_runtime;
REVOKE ALL ON FUNCTION system_resource_layout_installation_record_v2(
    uuid,uuid,uuid,uuid,uuid,timestamptz,name
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION system_resource_layout_installation_record_v2(
    uuid,uuid,uuid,uuid,uuid,timestamptz,name
) TO harpy_runtime;

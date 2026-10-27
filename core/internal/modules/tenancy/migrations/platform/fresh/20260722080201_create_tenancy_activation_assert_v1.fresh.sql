-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_assert_v1(
    p_operation_id uuid,
    p_tenant_id uuid,
    p_schema_name name
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_state text;
    v_kind text;
    v_tenant_id uuid;
    v_schema_name text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_tenant_id IS NULL
       OR p_schema_name IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant activation assertion denied';
    END IF;
    SELECT operation.state, operation.operation_kind,
           operation.tenant_id, operation.schema_name
    INTO v_state, v_kind, v_tenant_id, v_schema_name
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id = p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR v_state IS DISTINCT FROM 'finalizing'
       OR v_kind IS DISTINCT FROM 'environment_initialize'
       OR v_tenant_id IS DISTINCT FROM p_tenant_id
       OR v_schema_name IS DISTINCT FROM p_schema_name THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant activation assertion conflicts';
    END IF;
    PERFORM 1
    FROM platform.tenants AS tenant
    WHERE tenant.id = p_tenant_id
      AND tenant.schema_name = p_schema_name
      AND tenant.status = 'active'
    FOR KEY SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant activation target is not active';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_activation_assert_v1(
    uuid,uuid,name
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_assert_v1(
    uuid,uuid,name
) TO harpy_runtime;

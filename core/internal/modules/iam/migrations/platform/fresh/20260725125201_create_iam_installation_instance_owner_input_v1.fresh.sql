-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_instance_owner_input_v1(
    p_environment_id uuid
)
RETURNS TABLE (
    operation_id uuid,tenant_id uuid,schema_name name,user_id uuid,
    owner_binding_id uuid,request_sha256 text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR pg_catalog.current_setting('application_name',true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR p_environment_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='installation Instance Owner input denied';
    END IF;
    RETURN QUERY
    SELECT operation.operation_id,operation.tenant_id,operation.schema_name::name,
        operation.first_user_id,operation.instance_owner_binding_id,
        operation.step_plan_sha256
    FROM platform.tenancy_onboarding_operations AS operation
    JOIN platform.platformauthority_installation_owner_bindings AS authority
      ON authority.operation_id=operation.operation_id
     AND authority.environment_id=operation.environment_id
     AND authority.tenant_id=operation.tenant_id
     AND authority.instance_owner_binding_id=operation.instance_owner_binding_id
    WHERE operation.environment_id=p_environment_id
      AND operation.operation_kind='environment_initialize'
      AND operation.state='succeeded';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='completed installation Instance Owner input is unavailable';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_installation_instance_owner_input_v1(uuid)
FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_installation_instance_owner_input_v1(uuid)
TO forgecrud_runtime;

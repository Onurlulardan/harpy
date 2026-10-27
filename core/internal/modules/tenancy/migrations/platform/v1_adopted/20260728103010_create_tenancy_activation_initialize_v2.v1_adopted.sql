-- +goose Up
-- Migration stream version: 20260728103010.
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_initialize_v2(
    p_operation_id uuid,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_tenant_name text,
    p_tenant_slug text,
    p_owner_user_id uuid,
    p_schema_name text,
    p_binding_purpose text,
    p_manifest_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_replayed boolean;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'harpy-v2:core-operator-initialize' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant activation initialization v2 denied';
    END IF;
    SELECT platform.tenancy_activation_initialize_v1(
        p_operation_id, p_environment_id, p_tenant_id,
        p_tenant_name, p_tenant_slug, p_owner_user_id,
        p_schema_name, p_binding_purpose, p_manifest_sha256
    ) INTO v_replayed;
    PERFORM platform.cutover_initialize_tenant_v2(
        p_operation_id, p_tenant_id, p_owner_user_id, p_manifest_sha256
    );
    RETURN v_replayed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v2(
    uuid, uuid, uuid, text, text, uuid, text, text, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v2(
    uuid, uuid, uuid, text, text, uuid, text, text, text
) TO harpy_runtime;

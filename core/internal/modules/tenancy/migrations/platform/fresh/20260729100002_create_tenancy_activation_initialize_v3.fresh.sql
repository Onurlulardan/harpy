-- +goose Up
-- Migration stream version: 20260729100002.
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_initialize_v3(
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
    v_authority_id uuid;
    v_created_at timestamptz;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant activation initialization v3 denied';
    END IF;
    SELECT platform.tenancy_activation_initialize_v2(
        p_operation_id, p_environment_id, p_tenant_id,
        p_tenant_name, p_tenant_slug, p_owner_user_id,
        p_schema_name, p_binding_purpose, p_manifest_sha256
    ) INTO v_replayed;
    SELECT tenant.created_at INTO STRICT v_created_at
    FROM platform.tenants AS tenant
    WHERE tenant.id = p_tenant_id
      AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE', 'TRIAL');
    v_authority_id := pg_catalog.overlay(
        p_operation_id::text, '7', 15, 1
    )::uuid;
    PERFORM platform.workflow_register_tenant_authority_v1(
        v_authority_id, p_tenant_id, 1, 1, v_created_at
    );
    PERFORM platform.analytics_register_tenant_authority_v1(
        v_authority_id, p_tenant_id, 1, 1, v_created_at
    );
    RETURN v_replayed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v3(
    uuid, uuid, uuid, text, text, uuid, text, text, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v3(
    uuid, uuid, uuid, text, text, uuid, text, text, text
) TO forgecrud_runtime;

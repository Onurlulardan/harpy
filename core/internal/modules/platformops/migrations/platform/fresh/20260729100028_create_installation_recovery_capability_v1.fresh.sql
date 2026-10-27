-- +goose Up
-- Migration stream version: 20260729100028.
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_installation_recovery_initialize_v2(
    p_operation_id uuid,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_schema_name name,
    p_deployment_mode text,
    p_signing_generation bigint,
    p_trust_bundle_sha256 text,
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
    IF NOT platform.tenancy_activation_step_context_assert_v1(
        p_operation_id,
        'platformops.initialize_recovery_signing_identity',
        p_environment_id, p_tenant_id, p_schema_name,
        'environment_initialize'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'installation recovery capability required';
    END IF;
    v_replayed := platform.platformops_installation_recovery_initialize_v1(
        p_operation_id, p_environment_id, p_tenant_id, p_schema_name,
        p_deployment_mode, p_signing_generation, p_trust_bundle_sha256,
        p_manifest_sha256
    );
    IF NOT platform.tenancy_activation_step_consume_v1(
        p_operation_id,
        'platformops.initialize_recovery_signing_identity'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'installation recovery capability consume failed';
    END IF;
    RETURN v_replayed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platformops_installation_recovery_initialize_v2(
        uuid,uuid,uuid,name,text,bigint,text,text
    ) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.platformops_installation_recovery_initialize_v2(
        uuid,uuid,uuid,name,text,bigint,text,text
    ) TO forgecrud_runtime;

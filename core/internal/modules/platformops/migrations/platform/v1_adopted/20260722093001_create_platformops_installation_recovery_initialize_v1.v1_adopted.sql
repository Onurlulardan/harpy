-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platformops_installation_recovery_initialize_v1(
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
    v_existing platform.platformops_installation_recovery_identities%ROWTYPE;
    v_state text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name', true)
            IS DISTINCT FROM 'forgecrud-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_environment_id IS NULL
       OR p_tenant_id IS NULL OR p_schema_name IS NULL
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'installation recovery identity initialization denied';
    END IF;
    IF NOT platform.tenancy_activation_assert_v1(
        p_operation_id, p_tenant_id, p_schema_name
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'installation recovery identity activation conflicts';
    END IF;
    IF p_deployment_mode = 'onpremise' THEN
        IF p_signing_generation IS NULL OR p_signing_generation < 1
           OR p_trust_bundle_sha256 !~ '^[0-9a-f]{64}$' THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'on-premise recovery identity is invalid';
        END IF;
        v_state := 'active';
    ELSIF p_deployment_mode = 'saas' THEN
        IF p_signing_generation IS NOT NULL OR p_trust_bundle_sha256 IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'SaaS recovery identity must be not applicable';
        END IF;
        v_state := 'not_applicable';
    ELSE
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'deployment mode is invalid';
    END IF;
    SELECT * INTO v_existing
    FROM platform.platformops_installation_recovery_identities AS identity
    WHERE identity.operation_id = p_operation_id
    FOR SHARE;
    IF FOUND THEN
        IF v_existing.environment_id IS DISTINCT FROM p_environment_id
           OR v_existing.tenant_id IS DISTINCT FROM p_tenant_id
           OR v_existing.deployment_mode IS DISTINCT FROM p_deployment_mode
           OR v_existing.identity_state IS DISTINCT FROM v_state
           OR v_existing.signing_generation IS DISTINCT FROM p_signing_generation
           OR v_existing.trust_bundle_sha256 IS DISTINCT FROM p_trust_bundle_sha256
           OR v_existing.manifest_sha256 IS DISTINCT FROM p_manifest_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'installation recovery identity replay conflicts';
        END IF;
        RETURN true;
    END IF;
    INSERT INTO platform.platformops_installation_recovery_identities (
        operation_id, environment_id, tenant_id, deployment_mode,
        identity_state, signing_generation, trust_bundle_sha256,
        manifest_sha256, created_at
    ) VALUES (
        p_operation_id, p_environment_id, p_tenant_id, p_deployment_mode,
        v_state, p_signing_generation, p_trust_bundle_sha256,
        p_manifest_sha256, pg_catalog.statement_timestamp()
    );
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platformops_installation_recovery_initialize_v1(
    uuid,uuid,uuid,name,text,bigint,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platformops_installation_recovery_initialize_v1(
    uuid,uuid,uuid,name,text,bigint,text,text
) TO harpy_runtime;

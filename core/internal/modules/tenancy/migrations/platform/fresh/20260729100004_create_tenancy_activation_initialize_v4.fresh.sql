-- +goose Up
-- Migration stream version: 20260729100004.
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_initialize_v4(
    p_operation_id uuid,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_tenant_name text,
    p_tenant_slug text,
    p_owner_user_id uuid,
    p_schema_name text,
    p_binding_purpose text,
    p_manifest_sha256 text,
    p_operation_kind text
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_operation platform.tenancy_onboarding_operations%ROWTYPE;
    v_existing platform.tenancy_activation_facts%ROWTYPE;
    v_application text;
    v_now timestamptz;
    v_created_at timestamptz;
    v_authority_id uuid;
    v_replayed boolean;
BEGIN
    v_application := pg_catalog.current_setting('application_name', true);
    v_now := pg_catalog.statement_timestamp();
    v_replayed := false;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR NOT (
           (p_operation_kind = 'environment_initialize'
            AND v_application = 'harpy-v2:core-operator-initialize')
           OR (p_operation_kind = 'saas_customer_tenant_provision'
               AND v_application = 'harpy-v2:core-worker')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant activation initialization v4 denied';
    END IF;
    IF p_tenant_name IS NULL
       OR pg_catalog.char_length(p_tenant_name) NOT BETWEEN 1 AND 255
       OR p_tenant_name IS DISTINCT FROM pg_catalog.btrim(p_tenant_name)
       OR p_tenant_slug IS NULL
       OR pg_catalog.char_length(p_tenant_slug) NOT BETWEEN 1 AND 100
       OR p_tenant_slug IS DISTINCT FROM pg_catalog.lower(p_tenant_slug)
       OR p_schema_name !~ '^[a-z][a-z0-9_]*$'
       OR p_manifest_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'tenant activation input is invalid';
    END IF;
    SELECT * INTO v_operation
    FROM platform.tenancy_onboarding_operations AS operation
    WHERE operation.operation_id = p_operation_id
    FOR UPDATE;
    IF NOT FOUND OR v_operation.state IS DISTINCT FROM 'finalizing'
       OR v_operation.operation_kind IS DISTINCT FROM p_operation_kind
       OR v_operation.environment_id IS DISTINCT FROM p_environment_id
       OR v_operation.tenant_id IS DISTINCT FROM p_tenant_id
       OR v_operation.first_user_id IS DISTINCT FROM p_owner_user_id
       OR v_operation.schema_name::text IS DISTINCT FROM p_schema_name
       OR v_operation.binding_purpose IS DISTINCT FROM p_binding_purpose
       OR v_operation.manifest_sha256 IS DISTINCT FROM p_manifest_sha256
       OR NOT (
           (v_operation.operation_kind = 'environment_initialize'
            AND ((v_operation.deployment_mode = 'onpremise'
                  AND v_operation.binding_purpose = 'onpremise_primary')
                 OR (v_operation.deployment_mode = 'saas'
                     AND v_operation.binding_purpose = 'saas_control')))
           OR (v_operation.operation_kind = 'saas_customer_tenant_provision'
               AND v_operation.deployment_mode = 'saas'
               AND v_operation.binding_purpose = 'saas_customer')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant activation plan conflict';
    END IF;
    IF p_operation_kind = 'environment_initialize' THEN
        RETURN platform.tenancy_activation_initialize_v3(
            p_operation_id, p_environment_id, p_tenant_id,
            p_tenant_name, p_tenant_slug, p_owner_user_id,
            p_schema_name, p_binding_purpose, p_manifest_sha256
        );
    END IF;
    SELECT * INTO v_existing
    FROM platform.tenancy_activation_facts AS fact
    WHERE fact.operation_id = p_operation_id;
    IF FOUND THEN
        IF v_existing.environment_id IS DISTINCT FROM p_environment_id
           OR v_existing.tenant_id IS DISTINCT FROM p_tenant_id
           OR v_existing.owner_user_id IS DISTINCT FROM p_owner_user_id
           OR v_existing.binding_purpose IS DISTINCT FROM p_binding_purpose
           OR v_existing.manifest_sha256 IS DISTINCT FROM p_manifest_sha256 THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'tenant activation replay conflict';
        END IF;
        v_replayed := true;
    ELSE
        INSERT INTO platform.tenants (
            id, name, slug, status, settings, owner_id, schema_name,
            created_at, updated_at, deleted_at
        ) VALUES (
            p_tenant_id, p_tenant_name, p_tenant_slug, 'ACTIVE', '{}'::jsonb,
            p_owner_user_id, p_schema_name, v_now, v_now, NULL
        );
        INSERT INTO platform.tenancy_tenant_bindings (
            environment_id, tenant_id, purpose, created_at
        ) VALUES (p_environment_id, p_tenant_id, p_binding_purpose, v_now);
        INSERT INTO platform.tenancy_activation_audits (
            operation_id, environment_id, tenant_id, actor_id,
            event_type, manifest_sha256, recorded_at
        ) VALUES (
            p_operation_id, p_environment_id, p_tenant_id, p_owner_user_id,
            'installation.activated', p_manifest_sha256, v_now
        );
        INSERT INTO platform.tenancy_activation_facts (
            operation_id, environment_id, tenant_id, owner_user_id,
            binding_purpose, manifest_sha256, activated_at
        ) VALUES (
            p_operation_id, p_environment_id, p_tenant_id, p_owner_user_id,
            p_binding_purpose, p_manifest_sha256, v_now
        );
    END IF;
    PERFORM platform.cutover_initialize_tenant_v3(
        p_operation_id, p_tenant_id, p_owner_user_id,
        p_manifest_sha256, p_operation_kind
    );
    SELECT tenant.created_at INTO STRICT v_created_at
    FROM platform.tenants AS tenant
    WHERE tenant.id = p_tenant_id
      AND tenant.deleted_at IS NULL
      AND tenant.status IN ('ACTIVE', 'TRIAL');
    v_authority_id := pg_catalog.overlay(
        p_operation_id::text, '7', 15, 1
    )::uuid;
    PERFORM platform.analytics_register_tenant_authority_v1(
        v_authority_id, p_tenant_id, 1, 1, v_created_at
    );
    RETURN v_replayed;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v4(
    uuid,uuid,uuid,text,text,uuid,text,text,text,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v4(
    uuid,uuid,uuid,text,text,uuid,text,text,text,text
) TO harpy_runtime;

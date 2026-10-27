-- +goose Up
CREATE TABLE platform.tenancy_activation_facts (
    operation_id uuid PRIMARY KEY,
    environment_id uuid NOT NULL,
    tenant_id uuid NOT NULL UNIQUE,
    owner_user_id uuid NOT NULL,
    binding_purpose text NOT NULL,
    manifest_sha256 text NOT NULL CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'),
    activated_at timestamptz NOT NULL
);

CREATE TRIGGER tenancy_activation_facts_reject_mutation
BEFORE UPDATE OR DELETE ON platform.tenancy_activation_facts
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_reject_mutation();

-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_activation_initialize_v1(
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
    v_operation platform.tenancy_onboarding_operations%ROWTYPE;
    v_existing platform.tenancy_activation_facts%ROWTYPE;
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'tenant activation initialization denied';
    END IF;
    IF p_tenant_name IS NULL OR pg_catalog.char_length(p_tenant_name) NOT BETWEEN 1 AND 255
       OR p_tenant_name IS DISTINCT FROM pg_catalog.btrim(p_tenant_name)
       OR p_tenant_slug IS NULL OR pg_catalog.char_length(p_tenant_slug) NOT BETWEEN 1 AND 100
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
       OR v_operation.environment_id IS DISTINCT FROM p_environment_id
       OR v_operation.tenant_id IS DISTINCT FROM p_tenant_id
       OR v_operation.first_user_id IS DISTINCT FROM p_owner_user_id
       OR v_operation.schema_name IS DISTINCT FROM p_schema_name
       OR v_operation.binding_purpose IS DISTINCT FROM p_binding_purpose
       OR v_operation.manifest_sha256 IS DISTINCT FROM p_manifest_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant activation plan conflict';
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
        RETURN true;
    END IF;
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
    PERFORM platform.tenancy_activation_cutover_initialize_v1(
        p_operation_id, p_tenant_id, p_owner_user_id, p_manifest_sha256
    );
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
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.tenancy_activation_initialize_v1(
    uuid, uuid, uuid, text, text, uuid, text, text, text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.tenancy_activation_initialize_v1(
    uuid, uuid, uuid, text, text, uuid, text, text, text
) TO forgecrud_runtime;

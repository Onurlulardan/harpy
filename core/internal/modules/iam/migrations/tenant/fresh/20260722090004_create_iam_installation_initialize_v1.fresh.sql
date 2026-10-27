-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_installation_initialize_v1(
    p_operation_id uuid,
    p_environment_id uuid,
    p_tenant_id uuid,
    p_user_id uuid,
    p_organization_id uuid,
    p_protected_assignment_id uuid,
    p_email text,
    p_display_name text,
    p_tenant_name text,
    p_binding_set_sha256 text,
    p_user_schema_id uuid,
    p_user_field_id uuid,
    p_organization_schema_id uuid,
    p_organization_field_id uuid,
    p_role_schema_id uuid,
    p_role_field_id uuid,
    p_identifier_id uuid,
    p_access_mode_id uuid,
    p_membership_id uuid,
    p_role_id uuid,
    p_permission_id uuid,
    p_role_permission_id uuid
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_existing iam_installation_identity_facts%ROWTYPE;
    v_now timestamptz := pg_catalog.statement_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM installation initialization denied';
    END IF;
    IF p_email IS NULL OR p_email IS DISTINCT FROM pg_catalog.lower(p_email)
       OR pg_catalog.char_length(p_email) NOT BETWEEN 3 AND 254
       OR p_email !~ '^[^[:space:]@]+@[^[:space:]@]+$'
       OR p_display_name IS NULL OR pg_catalog.char_length(p_display_name) NOT BETWEEN 1 AND 128
       OR p_display_name IS DISTINCT FROM pg_catalog.btrim(p_display_name)
       OR p_tenant_name IS NULL OR pg_catalog.char_length(p_tenant_name) NOT BETWEEN 1 AND 128
       OR p_tenant_name IS DISTINCT FROM pg_catalog.btrim(p_tenant_name)
       OR p_binding_set_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM installation input is invalid';
    END IF;
    SELECT * INTO v_existing
    FROM iam_installation_identity_facts AS fact
    WHERE fact.operation_id = p_operation_id;
    IF FOUND THEN
        IF v_existing.environment_id IS DISTINCT FROM p_environment_id
           OR v_existing.tenant_id IS DISTINCT FROM p_tenant_id
           OR v_existing.first_user_id IS DISTINCT FROM p_user_id
           OR v_existing.root_organization_id IS DISTINCT FROM p_organization_id
           OR v_existing.protected_admin_binding_id IS DISTINCT FROM p_protected_assignment_id
           OR v_existing.normalized_email IS DISTINCT FROM p_email
           OR v_existing.protected_admin_binding_set_sha256 IS DISTINCT FROM
                pg_catalog.decode(p_binding_set_sha256, 'hex') THEN
            RAISE EXCEPTION USING ERRCODE = '23505',
                MESSAGE = 'IAM installation replay conflict';
        END IF;
        RETURN true;
    END IF;
    PERFORM iam_installation_identity_core_v1(
        p_operation_id, p_environment_id, p_tenant_id, p_user_id,
        p_organization_id, p_email, p_display_name, p_tenant_name,
        p_user_schema_id, p_user_field_id, p_organization_schema_id,
        p_organization_field_id, p_role_schema_id, p_role_field_id,
        p_identifier_id, p_access_mode_id, p_membership_id, v_now
    );
    PERFORM iam_installation_protected_admin_v1(
        p_tenant_id, p_user_id, p_protected_assignment_id, p_role_id,
        p_permission_id, p_role_permission_id, p_role_schema_id, v_now
    );
    INSERT INTO iam_installation_identity_facts (
        operation_id, environment_id, tenant_id, first_user_id,
        root_organization_id, protected_admin_binding_id,
        protected_admin_binding_set_sha256, normalized_email, activated_at
    ) VALUES (
        p_operation_id, p_environment_id, p_tenant_id, p_user_id,
        p_organization_id, p_protected_assignment_id,
        pg_catalog.decode(p_binding_set_sha256, 'hex'), p_email, v_now
    );
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_initialize_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_installation_initialize_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) TO harpy_runtime;

-- +goose Up
-- Migration stream version: 20260729100025.
-- +goose StatementBegin
CREATE FUNCTION iam_installation_initialize_v2(
    p_operation_id uuid, p_environment_id uuid, p_tenant_id uuid,
    p_user_id uuid, p_organization_id uuid, p_protected_assignment_id uuid,
    p_email text, p_display_name text, p_tenant_name text,
    p_binding_set_sha256 text, p_user_schema_id uuid, p_user_field_id uuid,
    p_organization_schema_id uuid, p_organization_field_id uuid,
    p_role_schema_id uuid, p_role_field_id uuid, p_identifier_id uuid,
    p_access_mode_id uuid, p_membership_id uuid, p_role_id uuid,
    p_permission_id uuid, p_role_permission_id uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF NOT platform.iam_installation_activation_assert_v2(
        p_operation_id, p_environment_id, p_tenant_id,
        (pg_catalog.current_schemas(false))[2]::name
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM installation identity capability required';
    END IF;
    RETURN iam_installation_initialize_v1(
        p_operation_id, p_environment_id, p_tenant_id, p_user_id,
        p_organization_id, p_protected_assignment_id, p_email,
        p_display_name, p_tenant_name, p_binding_set_sha256,
        p_user_schema_id, p_user_field_id, p_organization_schema_id,
        p_organization_field_id, p_role_schema_id, p_role_field_id,
        p_identifier_id, p_access_mode_id, p_membership_id, p_role_id,
        p_permission_id, p_role_permission_id
    );
END;
$function$;

CREATE FUNCTION iam_installation_credential_initialize_v2(
    p_scope_schema name, p_operation_id uuid, p_environment_id uuid,
    p_tenant_id uuid, p_user_id uuid, p_credential_id uuid,
    p_credential_version bigint, p_salt bytea, p_tag bytea
)
RETURNS TABLE (
    operation_id uuid, tenant_id uuid, user_id uuid, credential_id uuid,
    credential_version bigint, credential_revision bigint,
    marker_revision bigint
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
SET application_name TO 'harpy-v2:core-operator-initialize'
AS $function$
BEGIN
    IF NOT platform.tenancy_activation_step_context_assert_v1(
        p_operation_id, 'iam.initialize_tenant_identity',
        p_environment_id, p_tenant_id, p_scope_schema,
        'environment_initialize'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM installation credential capability required';
    END IF;
    RETURN QUERY SELECT * FROM iam_installation_credential_initialize_v1(
        p_scope_schema, p_operation_id, p_tenant_id, p_user_id,
        p_credential_id, p_credential_version, p_salt, p_tag
    );
    IF NOT platform.tenancy_activation_step_consume_v1(
        p_operation_id, 'iam.initialize_tenant_identity'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM installation credential capability consume failed';
    END IF;
END;
$function$;

CREATE FUNCTION iam_initial_admin_setup_finalize_v2(
    p_scope_schema name, p_operation_id uuid, p_environment_id uuid,
    p_tenant_id uuid, p_manifest_sha256 text
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
SET application_name TO 'harpy-v2:core-operator-initialize'
AS $function$
DECLARE
    v_artifact_id uuid;
BEGIN
    IF NOT platform.tenancy_activation_step_context_assert_v1(
        p_operation_id, 'iam.initialize_tenant_identity',
        p_environment_id, p_tenant_id, p_scope_schema,
        'saas_customer_tenant_provision'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'initial admin setup capability required';
    END IF;
    v_artifact_id := iam_initial_admin_setup_finalize_v1(
        p_scope_schema, p_operation_id, p_tenant_id, p_manifest_sha256
    );
    IF NOT platform.tenancy_activation_step_consume_v1(
        p_operation_id, 'iam.initialize_tenant_identity'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'initial admin setup capability consume failed';
    END IF;
    RETURN v_artifact_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_installation_initialize_v2(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_installation_initialize_v2(
    uuid,uuid,uuid,uuid,uuid,uuid,text,text,text,text,uuid,uuid,
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid
) TO harpy_runtime;
REVOKE ALL ON FUNCTION iam_installation_credential_initialize_v2(
    name,uuid,uuid,uuid,uuid,uuid,bigint,bytea,bytea
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_installation_credential_initialize_v2(
    name,uuid,uuid,uuid,uuid,uuid,bigint,bytea,bytea
) TO harpy_runtime;
REVOKE ALL ON FUNCTION iam_initial_admin_setup_finalize_v2(
    name,uuid,uuid,uuid,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION iam_initial_admin_setup_finalize_v2(
    name,uuid,uuid,uuid,text
) TO harpy_runtime;

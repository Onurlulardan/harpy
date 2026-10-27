-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_installation_instance_owner_bind_v1(
    p_operation_id uuid,p_environment_id uuid,p_tenant_id uuid,p_schema_name name,
    p_user_id uuid,p_subject_id uuid,p_owner_role_id uuid,
    p_role_permission_id uuid,p_owner_binding_id uuid,
    p_authorization_version bigint,p_request_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.iam_platform_mutation_receipts%ROWTYPE;
    v_permission_id uuid;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR pg_catalog.current_setting('application_name',true)
            IS DISTINCT FROM 'harpy-v2:core-operator-initialize'
       OR p_operation_id IS NULL OR p_environment_id IS NULL
       OR p_tenant_id IS NULL OR p_schema_name IS NULL OR p_user_id IS NULL
       OR p_subject_id IS NULL OR p_owner_role_id IS NULL
       OR p_role_permission_id IS NULL OR p_owner_binding_id IS NULL
       OR p_authorization_version < 1
       OR p_request_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='installation Instance Owner binding denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended('iam-installation-owner|'||p_environment_id::text,0));
    SELECT * INTO v_existing
    FROM platform.iam_platform_mutation_receipts AS receipt
    WHERE receipt.operation_id=p_operation_id FOR SHARE;
    IF FOUND THEN
        IF v_existing.environment_id IS DISTINCT FROM p_environment_id
           OR v_existing.mutation_kind IS DISTINCT FROM 'initial_owner'
           OR v_existing.request_sha256 IS DISTINCT FROM
                pg_catalog.decode(p_request_sha256,'hex')
           OR v_existing.actor_user_id IS DISTINCT FROM p_user_id
           OR v_existing.target_user_id IS DISTINCT FROM p_user_id
           OR v_existing.subject_id IS DISTINCT FROM p_subject_id
           OR v_existing.role_id IS DISTINCT FROM p_owner_role_id
           OR v_existing.subject_role_binding_id IS DISTINCT FROM p_owner_binding_id
           OR v_existing.role_permission_binding_id IS DISTINCT FROM p_role_permission_id
           OR v_existing.target_authorization_version IS DISTINCT FROM
                p_authorization_version THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='installation Instance Owner replay conflicts';
        END IF;
        RETURN true;
    END IF;
    PERFORM 1
    FROM platform.tenancy_onboarding_operations AS operation
    JOIN platform.tenants AS tenant ON tenant.id=operation.tenant_id
    JOIN platform.tenancy_tenant_bindings AS binding
      ON binding.environment_id=operation.environment_id
     AND binding.tenant_id=operation.tenant_id
    JOIN platform.platformauthority_installation_owner_bindings AS authority
      ON authority.operation_id=operation.operation_id
     AND authority.environment_id=operation.environment_id
     AND authority.tenant_id=operation.tenant_id
     AND authority.instance_owner_binding_id=p_owner_binding_id
    WHERE operation.operation_id=p_operation_id
      AND operation.environment_id=p_environment_id
      AND operation.tenant_id=p_tenant_id
      AND operation.schema_name=p_schema_name
      AND operation.first_user_id=p_user_id
      AND operation.instance_owner_binding_id=p_owner_binding_id
      AND operation.operation_kind='environment_initialize'
      AND operation.state IN ('finalizing','succeeded')
      AND tenant.schema_name=p_schema_name AND tenant.owner_id=p_user_id
      AND tenant.status='active' AND tenant.deleted_at IS NULL
      AND binding.purpose IN ('onpremise_primary','saas_control')
    FOR SHARE OF operation,tenant,binding,authority;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='installation Instance Owner activation evidence is unavailable';
    END IF;
    PERFORM 1 FROM platform.iam_login_directory AS directory
    WHERE directory.tenant_id=p_tenant_id AND directory.user_id=p_user_id
      AND directory.state='active' FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='installation Instance Owner has no active login mapping';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.iam_platform_owner_policies)
       OR EXISTS (SELECT 1 FROM platform.iam_platform_subjects
                  WHERE environment_id=p_environment_id) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='platform owner authority exists without installation receipt';
    END IF;
    SELECT permission.id INTO STRICT v_permission_id
    FROM platform.iam_platform_permissions AS permission
    WHERE permission.permission_key='environment_all'
      AND permission.match_kind='environment_all' AND permission.state='active'
    FOR SHARE;
    INSERT INTO platform.iam_platform_subjects(
        id,environment_id,primary_tenant_id,user_id,state,revision,created_at,updated_at
    ) VALUES (p_subject_id,p_environment_id,p_tenant_id,p_user_id,'active',1,v_now,v_now);
    INSERT INTO platform.iam_platform_roles(
        id,environment_id,role_kind,role_key,state,revision,created_at,updated_at
    ) VALUES (p_owner_role_id,p_environment_id,'instance_owner','instance_owner',
        'active',1,v_now,v_now);
    INSERT INTO platform.iam_platform_role_permissions(
        id,environment_id,role_id,permission_id,state,revision,created_at,updated_at
    ) VALUES (p_role_permission_id,p_environment_id,p_owner_role_id,v_permission_id,
        'active',1,v_now,v_now);
    INSERT INTO platform.iam_platform_subject_roles(
        id,environment_id,subject_id,role_id,state,revision,created_at,updated_at
    ) VALUES (p_owner_binding_id,p_environment_id,p_subject_id,p_owner_role_id,
        'active',1,v_now,v_now);
    INSERT INTO platform.iam_platform_owner_policies(
        environment_id,owner_binding_id,revision,created_at,updated_at
    ) VALUES (p_environment_id,p_owner_binding_id,1,v_now,v_now);
    INSERT INTO platform.iam_platform_authorization_clocks(
        environment_id,authorization_clock,revision,created_at,updated_at
    ) VALUES (p_environment_id,1,1,v_now,v_now);
    INSERT INTO platform.iam_platform_mutation_receipts VALUES (
        p_operation_id,p_environment_id,'initial_owner',
        pg_catalog.decode(p_request_sha256,'hex'),p_user_id,p_user_id,p_subject_id,
        p_owner_role_id,p_owner_binding_id,p_role_permission_id,v_permission_id,
        p_authorization_version,1,v_now
    );
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_installation_instance_owner_bind_v1(
    uuid,uuid,uuid,name,uuid,uuid,uuid,uuid,uuid,bigint,text
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.iam_installation_instance_owner_bind_v1(
    uuid,uuid,uuid,name,uuid,uuid,uuid,uuid,uuid,bigint,text
) TO harpy_runtime;

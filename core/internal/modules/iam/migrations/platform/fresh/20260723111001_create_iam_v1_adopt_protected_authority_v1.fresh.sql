-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.iam_v1_adopt_protected_authority_v1(
    p_operation_id uuid,p_environment_id uuid,p_primary_tenant_id uuid,p_user_id uuid,
    p_subject_id uuid,p_owner_role_id uuid,p_role_permission_id uuid,p_owner_binding_id uuid,
    p_authorization_version bigint,p_protected_mapping_sha256 text,p_plan_sha256 text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path TO pg_catalog
AS $function$
DECLARE
    v_existing platform.iam_v1_protected_authority_adoption_receipts%ROWTYPE;
    v_permission_id uuid;
    v_now timestamptz := pg_catalog.transaction_timestamp();
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR pg_catalog.current_setting('application_name',true)
            !~ '^forgecrud-cutover/[0-9a-f-]{36}$'
       OR p_operation_id IS NULL OR p_environment_id IS NULL
       OR p_primary_tenant_id IS NULL OR p_user_id IS NULL OR p_subject_id IS NULL
       OR p_owner_role_id IS NULL OR p_role_permission_id IS NULL
       OR p_owner_binding_id IS NULL
       OR p_authorization_version<1
       OR p_protected_mapping_sha256 !~ '^[0-9a-f]{64}$'
       OR p_plan_sha256 !~ '^[0-9a-f]{64}$' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='V1 protected authority adoption denied';
    END IF;
    PERFORM pg_catalog.pg_advisory_xact_lock(
        pg_catalog.hashtextextended('v1-iam-owner|'||p_environment_id::text,0));
    SELECT * INTO v_existing
      FROM platform.iam_v1_protected_authority_adoption_receipts
     WHERE environment_id=p_environment_id FOR SHARE;
    IF FOUND THEN
        IF v_existing.operation_id IS DISTINCT FROM p_operation_id
           OR v_existing.primary_tenant_id IS DISTINCT FROM p_primary_tenant_id
           OR v_existing.user_id IS DISTINCT FROM p_user_id
           OR v_existing.authorization_version IS DISTINCT FROM p_authorization_version
           OR v_existing.subject_id IS DISTINCT FROM p_subject_id
           OR v_existing.owner_role_id IS DISTINCT FROM p_owner_role_id
           OR v_existing.owner_binding_id IS DISTINCT FROM p_owner_binding_id
           OR v_existing.protected_mapping_sha256 IS DISTINCT FROM
                pg_catalog.decode(p_protected_mapping_sha256,'hex')
           OR v_existing.plan_sha256 IS DISTINCT FROM pg_catalog.decode(p_plan_sha256,'hex') THEN
            RAISE EXCEPTION USING ERRCODE='23505',
                MESSAGE='V1 protected authority replay conflicts';
        END IF;
        RETURN true;
    END IF;
    PERFORM 1 FROM platform.tenants AS tenant
      JOIN platform.tenancy_tenant_bindings AS binding
        ON binding.tenant_id=tenant.id AND binding.environment_id=p_environment_id
     WHERE tenant.id=p_primary_tenant_id AND tenant.owner_id=p_user_id
       AND tenant.deleted_at IS NULL
       AND binding.purpose IN ('onpremise_primary','saas_control')
     FOR SHARE OF tenant,binding;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 primary tenant owner mapping is unavailable';
    END IF;
    PERFORM 1 FROM platform.iam_login_directory AS directory
     WHERE directory.tenant_id=p_primary_tenant_id AND directory.user_id=p_user_id
       AND directory.state='active' FOR SHARE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='V1 protected account has no active canonical login mapping';
    END IF;
    IF EXISTS (SELECT 1 FROM platform.iam_platform_owner_policies)
       OR EXISTS (SELECT 1 FROM platform.iam_platform_subjects
                  WHERE environment_id=p_environment_id) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='platform owner authority already exists without adoption receipt';
    END IF;
    SELECT permission.id INTO STRICT v_permission_id
      FROM platform.iam_platform_permissions AS permission
     WHERE permission.permission_key='environment_all'
       AND permission.match_kind='environment_all' AND permission.state='active'
     FOR SHARE;
    INSERT INTO platform.iam_platform_subjects(
        id,environment_id,primary_tenant_id,user_id,state,revision,created_at,updated_at
    ) VALUES (p_subject_id,p_environment_id,p_primary_tenant_id,p_user_id,
        'active',1,v_now,v_now);
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
        pg_catalog.decode(p_plan_sha256,'hex'),p_user_id,p_user_id,p_subject_id,
        p_owner_role_id,p_owner_binding_id,p_role_permission_id,v_permission_id,
        p_authorization_version,1,v_now
    );
    INSERT INTO platform.iam_v1_protected_authority_adoption_receipts(
        operation_id,environment_id,primary_tenant_id,user_id,authorization_version,subject_id,
        owner_role_id,owner_binding_id,protected_mapping_sha256,plan_sha256,
        adopted_at,recorded_by
    ) VALUES (p_operation_id,p_environment_id,p_primary_tenant_id,p_user_id,
        p_authorization_version,p_subject_id,
        p_owner_role_id,p_owner_binding_id,
        pg_catalog.decode(p_protected_mapping_sha256,'hex'),
        pg_catalog.decode(p_plan_sha256,'hex'),v_now,SESSION_USER);
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.iam_v1_adopt_protected_authority_v1(
    uuid,uuid,uuid,uuid,uuid,uuid,uuid,uuid,bigint,text,text
) FROM PUBLIC;

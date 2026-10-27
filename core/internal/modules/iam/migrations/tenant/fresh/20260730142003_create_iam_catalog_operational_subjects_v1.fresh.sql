-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_operational_resource_v1(
    p_component_id uuid, p_resource_slug text, p_resource_kind text,
    p_name text, p_description text, p_occurred_at timestamptz
) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource iam_resources%ROWTYPE;
    v_owner_domain text;
BEGIN
    v_owner_domain := CASE p_resource_kind
        WHEN 'form' THEN 'forms'
        WHEN 'workflow' THEN 'workflow'
        WHEN 'module' THEN 'catalog'
        ELSE NULL
    END;
    IF v_owner_domain IS NULL OR p_resource_slug IS NULL OR
       p_name IS NULL OR p_description IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM catalog resource input is invalid';
    END IF;
    SELECT * INTO v_resource FROM iam_resources AS resource
    WHERE resource.resource_key=p_resource_slug FOR UPDATE;
    IF NOT FOUND THEN
        INSERT INTO iam_resources(
            id,resource_key,owner_domain,resource_kind,display_name,
            description,is_protected,state,revision,created_at,updated_at
        ) VALUES (
            pg_catalog.uuidv7(),p_resource_slug,v_owner_domain,
            p_resource_kind,p_name,NULLIF(p_description,''),
            false,'active',1,p_occurred_at,p_occurred_at
        ) RETURNING * INTO v_resource;
    ELSIF ROW(v_resource.owner_domain,v_resource.resource_kind,
              v_resource.display_name,COALESCE(v_resource.description,''),
              v_resource.is_protected,v_resource.state)
          IS DISTINCT FROM
          ROW(v_owner_domain,p_resource_kind,p_name,p_description,
              false,'active') THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='IAM catalog operational resource differs';
    END IF;
    INSERT INTO iam_catalog_operational_resources(
        component_id,resource_slug,resource_id
    ) VALUES (p_component_id,p_resource_slug,v_resource.id);
    RETURN v_resource.id;
END;
$function$;

CREATE FUNCTION iam_catalog_operational_role_v1(
    p_scope_schema name, p_role_name text,
    p_role_description text, p_occurred_at timestamptz
) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_tenant_id uuid;
    v_schema_id uuid;
    v_role iam_roles%ROWTYPE;
    v_profile jsonb;
BEGIN
    SELECT tenant.id INTO v_tenant_id
    FROM platform.tenants AS tenant
    WHERE tenant.schema_name=p_scope_schema::text
      AND tenant.status='ACTIVE' AND tenant.deleted_at IS NULL;
    SELECT role.id INTO v_schema_id
    FROM iam_profile_schema_versions AS role
    WHERE role.resource_kind='role' AND role.state='active'
    ORDER BY role.version DESC LIMIT 1;
    IF v_tenant_id IS NULL OR v_schema_id IS NULL OR
       p_role_name IS NULL OR p_role_description IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='IAM catalog role prerequisites are unavailable';
    END IF;
    SELECT * INTO v_role FROM iam_roles AS role
    WHERE role.scope_kind='tenant' AND
          role.name_key=pg_catalog.convert_to(p_role_name,'UTF8')
    FOR UPDATE;
    IF NOT FOUND THEN
        v_profile := pg_catalog.jsonb_build_object(
            'display_name',p_role_name
        );
        INSERT INTO iam_roles(
            id,scope_kind,scope_id,organization_id,name,name_key,
            description,status,is_protected,is_default,
            profile_resource_kind,profile_schema_version_id,
            profile_values,profile_sha256,revision,created_at,updated_at
        ) VALUES (
            pg_catalog.uuidv7(),'tenant',v_tenant_id,NULL,
            p_role_name,pg_catalog.convert_to(p_role_name,'UTF8'),
            NULLIF(p_role_description,''),'active',false,false,
            'role',v_schema_id,v_profile,
            pg_catalog.sha256(pg_catalog.convert_to(v_profile::text,'UTF8')),
            1,p_occurred_at,p_occurred_at
        ) RETURNING * INTO v_role;
    ELSIF v_role.status IS DISTINCT FROM 'active' OR
          v_role.is_protected IS DISTINCT FROM false OR
          v_role.scope_id IS DISTINCT FROM v_tenant_id THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='IAM catalog operational role differs';
    END IF;
    RETURN v_role.id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_catalog_operational_resource_v1(
    uuid,text,text,text,text,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_catalog_operational_role_v1(
    name,text,text,timestamptz
) FROM PUBLIC;

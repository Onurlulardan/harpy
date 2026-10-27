-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_operational_resource_v2(
    p_component_id uuid,p_package_id uuid,p_package_version text,
    p_release_id uuid,p_owner_generation bigint,p_content_sha256 text,
    p_resource_slug text,p_resource_kind text,
    p_name text,p_description text,p_occurred_at timestamptz
) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_resource iam_resources%ROWTYPE;
    v_owner_domain text;
    v_projected_id uuid;
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
            MESSAGE='IAM catalog version resource input is invalid';
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
    ELSIF ROW(
              v_resource.owner_domain,v_resource.resource_kind,
              v_resource.is_protected,v_resource.state
          ) IS DISTINCT FROM ROW(
              v_owner_domain,p_resource_kind,false,'active'::text
          ) THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='IAM catalog version operational resource differs';
    ELSIF ROW(
              v_resource.display_name,COALESCE(v_resource.description,'')
          ) IS DISTINCT FROM ROW(p_name,p_description) THEN
        UPDATE iam_resources
        SET display_name=p_name,description=NULLIF(p_description,''),
            revision=revision+1,updated_at=p_occurred_at
        WHERE id=v_resource.id
        RETURNING * INTO v_resource;
    END IF;
    INSERT INTO iam_catalog_version_operational_resources(
        component_id,package_id,package_version,release_id,
        owner_generation,content_sha256,resource_slug,resource_id
    ) VALUES (
        p_component_id,p_package_id,p_package_version,p_release_id,
        p_owner_generation,p_content_sha256,p_resource_slug,v_resource.id
    ) ON CONFLICT DO NOTHING;
    SELECT projected.resource_id INTO STRICT v_projected_id
    FROM iam_catalog_version_operational_resources AS projected
    WHERE projected.component_id=p_component_id
      AND projected.package_id=p_package_id
      AND projected.package_version=p_package_version
      AND projected.release_id=p_release_id
      AND projected.owner_generation=p_owner_generation
      AND projected.content_sha256=p_content_sha256
      AND projected.resource_slug=p_resource_slug;
    IF v_projected_id<>v_resource.id THEN
        RAISE EXCEPTION USING ERRCODE='23505',
            MESSAGE='IAM catalog version resource mapping differs';
    END IF;
    RETURN v_resource.id;
END;
$function$;

CREATE FUNCTION iam_catalog_operational_role_v2(
    p_scope_schema name,p_role_name text,
    p_role_description text,p_occurred_at timestamptz
) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_role_id uuid;
BEGIN
    v_role_id := iam_catalog_operational_role_v1(
        p_scope_schema,p_role_name,p_role_description,p_occurred_at
    );
    UPDATE iam_roles
    SET description=NULLIF(p_role_description,''),
        revision=revision+1,updated_at=p_occurred_at
    WHERE id=v_role_id
      AND COALESCE(description,'')<>p_role_description;
    RETURN v_role_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_catalog_operational_resource_v2(
    uuid,uuid,text,uuid,bigint,text,text,text,text,text,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION iam_catalog_operational_role_v2(
    name,text,text,timestamptz
) FROM PUBLIC;

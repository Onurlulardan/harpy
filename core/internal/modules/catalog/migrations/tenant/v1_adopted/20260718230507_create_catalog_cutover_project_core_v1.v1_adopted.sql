-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_v1_adoption_project_core_v1(
    p_kind text,p_record jsonb
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_kind NOT IN ('installation','binding','dependency') OR
       pg_catalog.jsonb_typeof(p_record) IS DISTINCT FROM 'object' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog V1 adoption core projection denied';
    END IF;
    IF p_kind='installation' THEN
        INSERT INTO catalog_installations(
          installation_id,package_id,active_version,state,reference_epoch,
          row_version,updated_at
        ) VALUES (
          (p_record->>'installation_id')::uuid,
          (p_record->>'package_id')::uuid,p_record->>'active_version',
          p_record->>'state',(p_record->>'reference_epoch')::bigint,
          (p_record->>'row_version')::bigint,
          (p_record->>'updated_at')::timestamptz
        ) ON CONFLICT DO NOTHING;
        PERFORM 1 FROM catalog_installations AS row
        WHERE row.installation_id=(p_record->>'installation_id')::uuid
          AND to_jsonb(row)=p_record;
    ELSIF p_kind='binding' THEN
        INSERT INTO catalog_component_bindings(
          installation_id,component_id,owner_module,owner_generation,
          owner_contract,content_sha256,state
        ) VALUES (
          (p_record->>'installation_id')::uuid,
          (p_record->>'component_id')::uuid,p_record->>'owner_module',
          (p_record->>'owner_generation')::bigint,p_record->>'owner_contract',
          p_record->>'content_sha256',p_record->>'state'
        ) ON CONFLICT DO NOTHING;
        PERFORM 1 FROM catalog_component_bindings AS row
        WHERE row.installation_id=(p_record->>'installation_id')::uuid
          AND row.component_id=(p_record->>'component_id')::uuid
          AND to_jsonb(row)=p_record;
    ELSE
        INSERT INTO catalog_dependency_resolutions(
          installation_id,dependency_package_id,exact_version,required,plan_sha256
        ) VALUES (
          (p_record->>'installation_id')::uuid,
          (p_record->>'dependency_package_id')::uuid,
          p_record->>'exact_version',(p_record->>'required')::boolean,
          p_record->>'plan_sha256'
        ) ON CONFLICT DO NOTHING;
        PERFORM 1 FROM catalog_dependency_resolutions AS row
        WHERE row.installation_id=(p_record->>'installation_id')::uuid
          AND row.dependency_package_id=(p_record->>'dependency_package_id')::uuid
          AND to_jsonb(row)=p_record;
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption core projection differs';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_v1_adoption_project_core_v1(text,jsonb) FROM PUBLIC;


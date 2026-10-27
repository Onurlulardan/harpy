-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_v1_adoption_project_menu_v1(
    p_kind text,p_record jsonb
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_kind NOT IN ('menu_definition','menu_version','menu_item') OR
       pg_catalog.jsonb_typeof(p_record) IS DISTINCT FROM 'object' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog V1 adoption menu projection denied';
    END IF;
    IF p_kind='menu_definition' THEN
        INSERT INTO catalog_menu_definitions(
          definition_id,revision,active_version_id,updated_at
        ) VALUES (
          (p_record->>'definition_id')::uuid,
          (p_record->>'revision')::bigint,NULL,
          (p_record->>'updated_at')::timestamptz
        ) ON CONFLICT DO NOTHING;
        PERFORM 1 FROM catalog_menu_definitions AS row
        WHERE row.definition_id=(p_record->>'definition_id')::uuid
          AND row.revision=(p_record->>'revision')::bigint
          AND row.active_version_id IS NULL
          AND row.updated_at=(p_record->>'updated_at')::timestamptz;
    ELSIF p_kind='menu_version' THEN
        INSERT INTO catalog_menu_versions(
          version_id,definition_id,version_number,content_sha256,created_at
        ) VALUES (
          (p_record->>'version_id')::uuid,
          (p_record->>'definition_id')::uuid,
          (p_record->>'version_number')::bigint,
          p_record->>'content_sha256',
          (p_record->>'created_at')::timestamptz
        ) ON CONFLICT DO NOTHING;
        PERFORM 1 FROM catalog_menu_versions AS row
        WHERE row.version_id=(p_record->>'version_id')::uuid
          AND to_jsonb(row)=p_record;
    ELSE
        INSERT INTO catalog_menu_items(
          version_id,item_id,parent_item_id,package_id,item_kind,target_id,
          external_url,label,sort_order
        ) VALUES (
          (p_record->>'version_id')::uuid,(p_record->>'item_id')::uuid,
          (p_record->>'parent_item_id')::uuid,(p_record->>'package_id')::uuid,
          p_record->>'item_kind',(p_record->>'target_id')::uuid,
          p_record->>'external_url',p_record->>'label',
          (p_record->>'sort_order')::integer
        ) ON CONFLICT DO NOTHING;
        PERFORM 1 FROM catalog_menu_items AS row
        WHERE row.version_id=(p_record->>'version_id')::uuid
          AND row.item_id=(p_record->>'item_id')::uuid
          AND to_jsonb(row)=p_record;
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption menu projection differs';
    END IF;
    RETURN true;
END;
$function$;

CREATE FUNCTION catalog_v1_adoption_activate_menu_v1(p_record jsonb)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       pg_catalog.jsonb_typeof(p_record) IS DISTINCT FROM 'object' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog V1 adoption menu activation denied';
    END IF;
    UPDATE catalog_menu_definitions
    SET active_version_id=(p_record->>'active_version_id')::uuid
    WHERE definition_id=(p_record->>'definition_id')::uuid
      AND active_version_id IS NULL
      AND p_record->>'active_version_id' IS NOT NULL;
    PERFORM 1 FROM catalog_menu_definitions AS row
    WHERE row.definition_id=(p_record->>'definition_id')::uuid
      AND to_jsonb(row)=p_record;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Catalog V1 adoption menu activation differs';
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_v1_adoption_project_menu_v1(text,jsonb) FROM PUBLIC;
REVOKE ALL ON FUNCTION catalog_v1_adoption_activate_menu_v1(jsonb) FROM PUBLIC;


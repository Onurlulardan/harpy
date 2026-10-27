-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_v1_adoption_value_valid_v1(
    p_key jsonb, p_value jsonb
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_source_kind text;
    v_source_identity text;
    v_target_kind text;
    v_target_identity text;
    v_record jsonb;
    v_keys jsonb;
    v_expected jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       pg_catalog.jsonb_typeof(p_key) IS DISTINCT FROM 'object' OR
       pg_catalog.jsonb_typeof(p_value) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(p_key) AS key)
       IS DISTINCT FROM ARRAY[
         'source_identity','source_kind','target_identity','target_kind'
       ]::text[] OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(p_value) AS key)
       IS DISTINCT FROM ARRAY[
         'source_document','source_identity','source_kind','source_sha256',
         'target_identity','target_kind','target_record'
       ]::text[] THEN
        RETURN false;
    END IF;
    v_source_kind:=p_value->>'source_kind';
    v_source_identity:=p_value->>'source_identity';
    v_target_kind:=p_value->>'target_kind';
    v_target_identity:=p_value->>'target_identity';
    v_record:=p_value->'target_record';
    IF p_key->>'source_kind' IS DISTINCT FROM v_source_kind OR
       p_key->>'source_identity' IS DISTINCT FROM v_source_identity OR
       p_key->>'target_kind' IS DISTINCT FROM v_target_kind OR
       p_key->>'target_identity' IS DISTINCT FROM v_target_identity OR
       v_source_kind NOT IN (
         'modules','module_dependencies','module_forms',
         'module_schema_versions','module_migration_runs','menu_configs'
       ) OR pg_catalog.octet_length(v_source_identity) NOT BETWEEN 1 AND 256 OR
       v_target_kind NOT IN (
         'discard','installation','binding','dependency',
         'menu_definition','menu_version','menu_item'
       ) OR pg_catalog.octet_length(v_target_identity) NOT BETWEEN 1 AND 256 OR
       p_value->>'source_sha256' !~ '^[0-9a-f]{64}$' OR
       pg_catalog.jsonb_typeof(p_value->'source_document') IS DISTINCT FROM 'object' OR
       pg_catalog.octet_length((p_value->'source_document')::text)>4194304 OR
       pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
         (p_value->'source_document')::text,'UTF8'
       )),'hex') IS DISTINCT FROM p_value->>'source_sha256' OR
       pg_catalog.octet_length(v_record::text)>4194304 THEN
        RETURN false;
    END IF;
    IF v_target_kind='discard' THEN
        RETURN v_record='null'::jsonb AND
          v_target_identity=v_source_kind||':'||v_source_identity;
    END IF;
    IF pg_catalog.jsonb_typeof(v_record) IS DISTINCT FROM 'object' THEN
        RETURN false;
    END IF;
    SELECT pg_catalog.jsonb_agg(key ORDER BY key) INTO v_keys
    FROM pg_catalog.jsonb_object_keys(v_record) AS key;
    v_expected:=CASE v_target_kind
      WHEN 'installation' THEN
        '["active_version","installation_id","package_id","reference_epoch","row_version","state","updated_at"]'::jsonb
      WHEN 'binding' THEN
        '["component_id","content_sha256","installation_id","owner_contract","owner_generation","owner_module","state"]'::jsonb
      WHEN 'dependency' THEN
        '["dependency_package_id","exact_version","installation_id","plan_sha256","required"]'::jsonb
      WHEN 'menu_definition' THEN
        '["active_version_id","definition_id","revision","updated_at"]'::jsonb
      WHEN 'menu_version' THEN
        '["content_sha256","created_at","definition_id","version_id","version_number"]'::jsonb
      WHEN 'menu_item' THEN
        '["external_url","item_id","item_kind","label","package_id","parent_item_id","sort_order","target_id","version_id"]'::jsonb
    END;
    IF v_keys IS DISTINCT FROM v_expected THEN
        RETURN false;
    END IF;
    RETURN v_target_identity=CASE v_target_kind
      WHEN 'installation' THEN v_record->>'installation_id'
      WHEN 'binding' THEN v_record->>'component_id'
      WHEN 'dependency' THEN
        (v_record->>'installation_id')||':'||
        (v_record->>'dependency_package_id')
      WHEN 'menu_definition' THEN v_record->>'definition_id'
      WHEN 'menu_version' THEN v_record->>'version_id'
      WHEN 'menu_item' THEN
        (v_record->>'version_id')||':'||(v_record->>'item_id')
    END;
EXCEPTION WHEN invalid_text_representation OR numeric_value_out_of_range THEN
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION catalog_v1_adoption_value_valid_v1(jsonb,jsonb) FROM PUBLIC;

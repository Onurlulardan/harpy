-- +goose Up
-- Migration stream version: 20260802232100.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_curated_definition_v2(
    p_resource_kind text, p_catalog_json text
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_catalog jsonb; v_surface text; v_allowed jsonb; v_fields jsonb;
    v_surfaces jsonb; v_en text; v_tr text;
BEGIN
    v_surfaces := '[]'::jsonb;
    v_catalog := p_catalog_json::jsonb;
    IF p_resource_kind NOT IN ('user','organization','role')
       OR v_catalog ->> 'contract' <> 'forms-presentation-field-catalog.v1'
       OR pg_catalog.jsonb_typeof(v_catalog -> 'fields') <> 'array' THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'system resource curated definition v2 input is invalid';
    END IF;
    FOREACH v_surface IN ARRAY ARRAY[
        'record_create','record_update','record_detail',
        'record_list','record_filter'
    ] LOOP
        v_allowed := CASE p_resource_kind || ':' || v_surface
        WHEN 'user:record_create' THEN pg_catalog.to_jsonb(ARRAY[
            'a1000000-0000-4000-8000-000000000002',
            'a1000000-0000-4000-8000-000000000007',
            'a1000000-0000-4000-8000-000000000006'])
        WHEN 'user:record_update' THEN pg_catalog.to_jsonb(ARRAY[
            'a1000000-0000-4000-8000-000000000007',
            'a1000000-0000-4000-8000-000000000006'])
        WHEN 'user:record_detail' THEN pg_catalog.to_jsonb(ARRAY[
            'a1000000-0000-4000-8000-000000000002','a1000000-0000-4000-8000-000000000004',
            'a1000000-0000-4000-8000-000000000005','a1000000-0000-4000-8000-000000000007',
            'a1000000-0000-4000-8000-000000000006','a1000000-0000-4000-8000-00000000000a',
            'a1000000-0000-4000-8000-00000000000b'])
        WHEN 'user:record_list' THEN pg_catalog.to_jsonb(ARRAY[
            'a1000000-0000-4000-8000-000000000002','a1000000-0000-4000-8000-000000000004',
            'a1000000-0000-4000-8000-000000000005','a1000000-0000-4000-8000-000000000007',
            'a1000000-0000-4000-8000-000000000006','a1000000-0000-4000-8000-00000000000a'])
        WHEN 'user:record_filter' THEN pg_catalog.to_jsonb(ARRAY[
            'a1000000-0000-4000-8000-000000000002','a1000000-0000-4000-8000-000000000004',
            'a1000000-0000-4000-8000-000000000005'])
        WHEN 'role:record_create' THEN pg_catalog.to_jsonb(ARRAY[
            'b1000000-0000-4000-8000-000000000005','b1000000-0000-4000-8000-000000000006',
            'b1000000-0000-4000-8000-000000000002','b1000000-0000-4000-8000-000000000004',
            'b1000000-0000-4000-8000-000000000009'])
        WHEN 'role:record_update' THEN pg_catalog.to_jsonb(ARRAY[
            'b1000000-0000-4000-8000-000000000005','b1000000-0000-4000-8000-000000000006',
            'b1000000-0000-4000-8000-000000000009'])
        WHEN 'role:record_detail' THEN pg_catalog.to_jsonb(ARRAY[
            'b1000000-0000-4000-8000-000000000005','b1000000-0000-4000-8000-000000000006',
            'b1000000-0000-4000-8000-000000000007','b1000000-0000-4000-8000-000000000002',
            'b1000000-0000-4000-8000-000000000004','b1000000-0000-4000-8000-000000000008',
            'b1000000-0000-4000-8000-000000000009','b1000000-0000-4000-8000-00000000000d',
            'b1000000-0000-4000-8000-00000000000e'])
        WHEN 'role:record_list' THEN pg_catalog.to_jsonb(ARRAY[
            'b1000000-0000-4000-8000-000000000005','b1000000-0000-4000-8000-000000000006',
            'b1000000-0000-4000-8000-000000000007','b1000000-0000-4000-8000-000000000002',
            'b1000000-0000-4000-8000-000000000004','b1000000-0000-4000-8000-000000000008',
            'b1000000-0000-4000-8000-000000000009','b1000000-0000-4000-8000-00000000000d'])
        WHEN 'role:record_filter' THEN pg_catalog.to_jsonb(ARRAY[
            'b1000000-0000-4000-8000-000000000005','b1000000-0000-4000-8000-000000000007',
            'b1000000-0000-4000-8000-000000000002','b1000000-0000-4000-8000-000000000008',
            'b1000000-0000-4000-8000-000000000009'])
        WHEN 'organization:record_create' THEN pg_catalog.to_jsonb(ARRAY[
            'c1000000-0000-4000-8000-000000000002','c1000000-0000-4000-8000-000000000003',
            'c1000000-0000-4000-8000-000000000005','c1000000-0000-4000-8000-00000000000c'])
        WHEN 'organization:record_update' THEN pg_catalog.to_jsonb(ARRAY[
            'c1000000-0000-4000-8000-000000000002','c1000000-0000-4000-8000-000000000003',
            'c1000000-0000-4000-8000-000000000005','c1000000-0000-4000-8000-00000000000c'])
        WHEN 'organization:record_detail' THEN pg_catalog.to_jsonb(ARRAY[
            'c1000000-0000-4000-8000-000000000002','c1000000-0000-4000-8000-000000000003',
            'c1000000-0000-4000-8000-000000000004','c1000000-0000-4000-8000-000000000005',
            'c1000000-0000-4000-8000-00000000000c','c1000000-0000-4000-8000-00000000000a',
            'c1000000-0000-4000-8000-00000000000b'])
        WHEN 'organization:record_list' THEN pg_catalog.to_jsonb(ARRAY[
            'c1000000-0000-4000-8000-000000000002','c1000000-0000-4000-8000-000000000003',
            'c1000000-0000-4000-8000-000000000004','c1000000-0000-4000-8000-000000000005',
            'c1000000-0000-4000-8000-00000000000c','c1000000-0000-4000-8000-00000000000a'])
        ELSE pg_catalog.to_jsonb(ARRAY[
            'c1000000-0000-4000-8000-000000000002','c1000000-0000-4000-8000-000000000003',
            'c1000000-0000-4000-8000-000000000004']) END;
        WITH allowed AS (
            SELECT allowed.value AS id, allowed.ordinality AS allowed_order
            FROM pg_catalog.jsonb_array_elements_text(v_allowed)
                WITH ORDINALITY AS allowed(value, ordinality)
        ), selected AS (
            SELECT field.value ->> 'id' AS id,
                   CASE WHEN p_resource_kind = 'user' AND
                       field.value ->> 'id' NOT LIKE 'a1000000-%'
                       THEN 0 ELSE 1 END AS group_order,
                   allowed.allowed_order
            FROM pg_catalog.jsonb_array_elements(v_catalog -> 'fields') AS field(value)
            LEFT JOIN allowed ON allowed.id = field.value ->> 'id'
            WHERE allowed.id IS NOT NULL OR p_resource_kind = 'user'
               AND field.value ->> 'id' NOT LIKE 'a1000000-%'
        ), ordered AS (
            SELECT id, pg_catalog.row_number() OVER (
                ORDER BY group_order, allowed_order NULLS FIRST, id
            ) - 1 AS field_order FROM selected
        )
        SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
            'field_id',id,'interaction','auto','order',field_order,
            'translations',pg_catalog.jsonb_build_array(
                pg_catalog.jsonb_build_object('help','','label',
                    system_resource_layout_default_field_label_v1(id,'en'),'locale','en'),
                pg_catalog.jsonb_build_object('help','','label',
                    system_resource_layout_default_field_label_v1(id,'tr'),'locale','tr')),
            'widget','auto') ORDER BY field_order), '[]'::jsonb)
        INTO v_fields FROM ordered;
        IF pg_catalog.jsonb_array_length(v_fields) = 0 THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'system resource curated v2 surface has no fields';
        END IF;
        v_en := CASE v_surface WHEN 'record_create' THEN 'Create'
            WHEN 'record_update' THEN 'Update' WHEN 'record_detail' THEN 'Details'
            WHEN 'record_list' THEN 'List' ELSE 'Filters' END;
        v_tr := CASE v_surface WHEN 'record_create' THEN 'Oluştur'
            WHEN 'record_update' THEN 'Güncelle' WHEN 'record_detail' THEN 'Ayrıntılar'
            WHEN 'record_list' THEN 'Liste' ELSE 'Filtreler' END;
        v_surfaces := v_surfaces || pg_catalog.jsonb_build_array(
            pg_catalog.jsonb_build_object('kind',v_surface,'sections',
                pg_catalog.jsonb_build_array(pg_catalog.jsonb_build_object(
                    'fields',v_fields,'id','00000000-0000-4000-8000-000000000001',
                    'order',0,'translations',pg_catalog.jsonb_build_array(
                        pg_catalog.jsonb_build_object('help','','label','Fields','locale','en'),
                        pg_catalog.jsonb_build_object('help','','label','Alanlar','locale','tr')))),
                'translations',pg_catalog.jsonb_build_array(
                    pg_catalog.jsonb_build_object('help','','label',v_en,'locale','en'),
                    pg_catalog.jsonb_build_object('help','','label',v_tr,'locale','tr'))));
    END LOOP;
    RETURN system_resource_layout_canonical_json_v1(
        pg_catalog.jsonb_build_object(
            'contract','forgecrud/system-resource-layout/v2','surfaces',v_surfaces));
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_curated_definition_v2(text,text)
    FROM PUBLIC;

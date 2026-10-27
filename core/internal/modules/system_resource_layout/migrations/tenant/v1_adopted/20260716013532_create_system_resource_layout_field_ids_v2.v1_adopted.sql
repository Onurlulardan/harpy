-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_field_ids_v2(p_definition text)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
    v_result jsonb;
BEGIN
    IF NOT system_resource_layout_definition_supported_v2(p_definition) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_field_ids_v2_contract_guard',
            MESSAGE = 'system resource layout definition is invalid';
    END IF;
    v_document := p_definition::jsonb;
    IF v_document ->> 'contract' = 'forgecrud/system-resource-layout/v1' THEN
        SELECT COALESCE(
            pg_catalog.jsonb_agg(pg_catalog.to_jsonb(source.field_id)
                ORDER BY source.field_id),
            '[]'::jsonb
        ) INTO v_result
        FROM (
            SELECT DISTINCT field.value ->> 'field_id' AS field_id
            FROM pg_catalog.jsonb_array_elements(v_document -> 'sections')
                AS section(value)
            CROSS JOIN pg_catalog.jsonb_array_elements(section.value -> 'fields')
                AS field(value)
        ) AS source;
    ELSE
        SELECT COALESCE(
            pg_catalog.jsonb_agg(pg_catalog.to_jsonb(source.field_id)
                ORDER BY source.field_id),
            '[]'::jsonb
        ) INTO v_result
        FROM (
            SELECT DISTINCT field.value ->> 'field_id' AS field_id
            FROM pg_catalog.jsonb_array_elements(v_document -> 'surfaces')
                AS surface(value)
            CROSS JOIN pg_catalog.jsonb_array_elements(surface.value -> 'sections')
                AS section(value)
            CROSS JOIN pg_catalog.jsonb_array_elements(section.value -> 'fields')
                AS field(value)
        ) AS source;
    END IF;
    RETURN v_result;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_field_ids_v2(text)
    FROM PUBLIC;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_definition_valid_v2(p_definition text)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
BEGIN
    BEGIN
        v_document := p_definition::jsonb;
        IF pg_catalog.jsonb_typeof(v_document) IS DISTINCT FROM 'object'
           OR (SELECT pg_catalog.count(*)
               FROM pg_catalog.jsonb_object_keys(v_document)) <> 2
           OR EXISTS (
                SELECT 1 FROM pg_catalog.jsonb_object_keys(v_document) AS present(key)
                WHERE present.key NOT IN ('contract', 'surfaces')
           )
           OR v_document ->> 'contract' IS DISTINCT FROM
              'harpy/system-resource-layout/v2'
           OR pg_catalog.jsonb_typeof(v_document -> 'surfaces') IS DISTINCT FROM 'array'
        THEN
            RETURN false;
        END IF;
    EXCEPTION WHEN OTHERS THEN
        RETURN false;
    END;
    IF pg_catalog.jsonb_array_length(v_document -> 'surfaces') NOT BETWEEN 5 AND 6
    THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(v_document -> 'surfaces')
            WITH ORDINALITY AS surface(value, position)
        WHERE pg_catalog.jsonb_typeof(surface.value) IS DISTINCT FROM 'object'
    ) THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(v_document -> 'surfaces')
            WITH ORDINALITY AS surface(value, position)
        WHERE (SELECT pg_catalog.count(*)
               FROM pg_catalog.jsonb_object_keys(surface.value)) <> 3
           OR EXISTS (
                SELECT 1
                FROM pg_catalog.jsonb_object_keys(surface.value) AS present(key)
                WHERE present.key NOT IN ('kind', 'sections', 'translations')
           )
           OR pg_catalog.jsonb_typeof(surface.value -> 'kind') IS DISTINCT FROM 'string'
           OR surface.value ->> 'kind' IS DISTINCT FROM CASE surface.position
                WHEN 1 THEN 'record_create'
                WHEN 2 THEN 'record_update'
                WHEN 3 THEN 'record_detail'
                WHEN 4 THEN 'record_list'
                WHEN 5 THEN 'record_filter'
                WHEN 6 THEN 'print'
                ELSE ''
              END
           OR pg_catalog.jsonb_typeof(surface.value -> 'sections') IS DISTINCT FROM 'array'
           OR pg_catalog.jsonb_typeof(surface.value -> 'translations') IS DISTINCT FROM 'array'
    ) THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(v_document -> 'surfaces') AS surface(value)
        WHERE NOT system_resource_layout_translations_valid_v2(
                surface.value -> 'translations'
              )
           OR NOT system_resource_layout_sections_valid_v2(
                surface.value -> 'sections'
              )
    ) THEN
        RETURN false;
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_definition_valid_v2(text)
    FROM PUBLIC;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_sections_valid_v2(p_sections jsonb)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    IF pg_catalog.jsonb_typeof(p_sections) IS DISTINCT FROM 'array' THEN
        RETURN false;
    END IF;
    IF pg_catalog.jsonb_array_length(p_sections) > 64 THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(p_sections)
            WITH ORDINALITY AS section(value, position)
        WHERE pg_catalog.jsonb_typeof(section.value) IS DISTINCT FROM 'object'
           OR (SELECT pg_catalog.count(*)
               FROM pg_catalog.jsonb_object_keys(section.value)) <> 4
           OR EXISTS (
                SELECT 1
                FROM pg_catalog.jsonb_object_keys(section.value) AS present(key)
                WHERE present.key NOT IN ('fields', 'id', 'order', 'translations')
           )
           OR pg_catalog.jsonb_typeof(section.value -> 'id') IS DISTINCT FROM 'string'
           OR section.value ->> 'id' !~
              '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
           OR section.value ->> 'id' = '00000000-0000-0000-0000-000000000000'
           OR pg_catalog.jsonb_typeof(section.value -> 'order') IS DISTINCT FROM 'number'
           OR section.value ->> 'order' IS DISTINCT FROM
              (section.position - 1)::text
           OR pg_catalog.jsonb_typeof(section.value -> 'fields') IS DISTINCT FROM 'array'
           OR pg_catalog.jsonb_typeof(section.value -> 'translations') IS DISTINCT FROM 'array'
    ) THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_array_elements(p_sections) AS section(value)
        WHERE NOT system_resource_layout_translations_valid_v2(
            section.value -> 'translations'
        )
    ) THEN
        RETURN false;
    END IF;
    IF (SELECT pg_catalog.count(*)
        FROM pg_catalog.jsonb_array_elements(p_sections) AS section(value)
        CROSS JOIN pg_catalog.jsonb_array_elements(section.value -> 'fields')) > 512 THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(p_sections) AS section(value)
        CROSS JOIN pg_catalog.jsonb_array_elements(section.value -> 'fields')
            WITH ORDINALITY AS field(value, position)
        WHERE pg_catalog.jsonb_typeof(field.value) IS DISTINCT FROM 'object'
           OR (SELECT pg_catalog.count(*)
               FROM pg_catalog.jsonb_object_keys(field.value)) <> 5
           OR EXISTS (
                SELECT 1
                FROM pg_catalog.jsonb_object_keys(field.value) AS present(key)
                WHERE present.key NOT IN (
                    'field_id', 'interaction', 'order', 'translations', 'widget'
                )
           )
           OR pg_catalog.jsonb_typeof(field.value -> 'field_id') IS DISTINCT FROM 'string'
           OR field.value ->> 'field_id' !~
              '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
           OR field.value ->> 'field_id' = '00000000-0000-0000-0000-000000000000'
           OR pg_catalog.jsonb_typeof(field.value -> 'interaction') IS DISTINCT FROM 'string'
           OR field.value ->> 'interaction' NOT IN ('auto', 'read_only', 'disabled')
           OR pg_catalog.jsonb_typeof(field.value -> 'order') IS DISTINCT FROM 'number'
           OR field.value ->> 'order' IS DISTINCT FROM (field.position - 1)::text
           OR pg_catalog.jsonb_typeof(field.value -> 'translations') IS DISTINCT FROM 'array'
           OR pg_catalog.jsonb_typeof(field.value -> 'widget') IS DISTINCT FROM 'string'
           OR field.value ->> 'widget' NOT IN (
                'auto', 'checkbox', 'date', 'datetime', 'document_selector',
                'document_uploader', 'email', 'json', 'lookup', 'multi_select',
                'number', 'password', 'radio', 'select', 'telephone', 'text',
                'textarea', 'time', 'toggle', 'url'
           )
    ) THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(p_sections) AS section(value)
        CROSS JOIN pg_catalog.jsonb_array_elements(section.value -> 'fields')
            AS field(value)
        WHERE NOT system_resource_layout_translations_valid_v2(
            field.value -> 'translations'
        )
    ) OR EXISTS (
        SELECT section.value ->> 'id'
        FROM pg_catalog.jsonb_array_elements(p_sections) AS section(value)
        GROUP BY section.value ->> 'id' HAVING pg_catalog.count(*) > 1
    ) OR EXISTS (
        SELECT field.value ->> 'field_id'
        FROM pg_catalog.jsonb_array_elements(p_sections) AS section(value)
        CROSS JOIN pg_catalog.jsonb_array_elements(section.value -> 'fields')
            AS field(value)
        GROUP BY field.value ->> 'field_id' HAVING pg_catalog.count(*) > 1
    ) THEN
        RETURN false;
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_sections_valid_v2(jsonb)
    FROM PUBLIC;

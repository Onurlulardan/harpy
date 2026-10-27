-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_definition_valid_v1(p_definition text)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_document jsonb;
BEGIN
    BEGIN
        v_document := p_definition::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RETURN false;
    END;
    IF pg_catalog.jsonb_typeof(v_document) IS DISTINCT FROM 'object'
       OR (SELECT pg_catalog.count(*) FROM
           pg_catalog.jsonb_object_keys(v_document)) <> 2
       OR EXISTS (
           SELECT 1 FROM pg_catalog.jsonb_object_keys(v_document) AS root(key)
           WHERE root.key NOT IN ('contract', 'sections')
       )
       OR v_document ->> 'contract' IS DISTINCT FROM
          'harpy/system-resource-layout/v1'
       OR pg_catalog.jsonb_typeof(v_document -> 'sections') <> 'array'
       OR pg_catalog.jsonb_array_length(v_document -> 'sections') > 64 THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(v_document -> 'sections')
            AS section(value)
        WHERE pg_catalog.jsonb_typeof(section.value) <> 'object'
           OR (SELECT pg_catalog.count(*) FROM
               pg_catalog.jsonb_object_keys(section.value)) <> 4
           OR EXISTS (
                SELECT 1
                FROM pg_catalog.jsonb_object_keys(section.value) AS present(key)
                WHERE present.key NOT IN ('fields', 'id', 'order', 'translations')
           )
           OR pg_catalog.jsonb_typeof(section.value -> 'id') <> 'string'
           OR section.value ->> 'id' !~
              '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
           OR section.value ->> 'id' =
              '00000000-0000-0000-0000-000000000000'
           OR pg_catalog.jsonb_typeof(section.value -> 'order') <> 'number'
           OR section.value ->> 'order' !~ '^(0|[1-9][0-9]{0,8})$'
           OR pg_catalog.jsonb_typeof(section.value -> 'fields') <> 'array'
           OR pg_catalog.jsonb_typeof(section.value -> 'translations') <> 'array'
           OR pg_catalog.jsonb_array_length(section.value -> 'translations') > 32
    ) THEN
        RETURN false;
    END IF;
    IF (SELECT pg_catalog.count(*)
        FROM pg_catalog.jsonb_array_elements(v_document -> 'sections') AS section
        CROSS JOIN pg_catalog.jsonb_array_elements(
            section.value -> 'fields'
        )) > 512 OR EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(v_document -> 'sections')
            AS section(value)
        CROSS JOIN pg_catalog.jsonb_array_elements(
            section.value -> 'fields'
        ) AS field(value)
        WHERE pg_catalog.jsonb_typeof(field.value) <> 'object'
           OR (SELECT pg_catalog.count(*) FROM
               pg_catalog.jsonb_object_keys(field.value)) <> 4
           OR EXISTS (
                SELECT 1
                FROM pg_catalog.jsonb_object_keys(field.value) AS present(key)
                WHERE present.key NOT IN (
                    'field_id', 'order', 'translations', 'widget'
                )
           )
           OR pg_catalog.jsonb_typeof(field.value -> 'field_id') <> 'string'
           OR field.value ->> 'field_id' !~
              '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
           OR field.value ->> 'field_id' =
              '00000000-0000-0000-0000-000000000000'
           OR pg_catalog.jsonb_typeof(field.value -> 'order') <> 'number'
           OR field.value ->> 'order' !~ '^(0|[1-9][0-9]{0,8})$'
           OR pg_catalog.jsonb_typeof(field.value -> 'widget') <> 'string'
           OR field.value ->> 'widget' NOT IN (
                'auto', 'text', 'textarea', 'number', 'toggle',
                'select', 'date', 'datetime', 'json'
           )
           OR pg_catalog.jsonb_typeof(field.value -> 'translations') <> 'array'
           OR pg_catalog.jsonb_array_length(field.value -> 'translations') > 32
    ) THEN
        RETURN false;
    END IF;
    IF EXISTS (
        WITH translations AS (
            SELECT translation.value
            FROM pg_catalog.jsonb_array_elements(v_document -> 'sections')
                AS section(value)
            CROSS JOIN pg_catalog.jsonb_array_elements(
                section.value -> 'translations'
            ) AS translation(value)
            UNION ALL
            SELECT translation.value
            FROM pg_catalog.jsonb_array_elements(v_document -> 'sections')
                AS section(value)
            CROSS JOIN pg_catalog.jsonb_array_elements(
                section.value -> 'fields'
            ) AS field(value)
            CROSS JOIN pg_catalog.jsonb_array_elements(
                field.value -> 'translations'
            ) AS translation(value)
        )
        SELECT 1 FROM translations
        WHERE pg_catalog.jsonb_typeof(value) <> 'object'
           OR (SELECT pg_catalog.count(*) FROM
               pg_catalog.jsonb_object_keys(value)) <> 3
           OR EXISTS (
                SELECT 1 FROM pg_catalog.jsonb_object_keys(value) AS present(key)
                WHERE present.key NOT IN ('help', 'label', 'locale')
           )
           OR pg_catalog.jsonb_typeof(value -> 'help') <> 'string'
           OR pg_catalog.jsonb_typeof(value -> 'label') <> 'string'
           OR pg_catalog.jsonb_typeof(value -> 'locale') <> 'string'
           OR value ->> 'locale' !~
              '^[a-z]{2,3}(-[A-Z][a-z]{3})?(-([A-Z]{2}|[0-9]{3}))?$'
           OR pg_catalog.char_length(value ->> 'label') NOT BETWEEN 1 AND 200
           OR pg_catalog.char_length(value ->> 'help') > 1000
           OR value ->> 'label' ~ '[[:cntrl:]]'
           OR value ->> 'help' ~ '[[:cntrl:]]'
    ) THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT section.value ->> 'id'
        FROM pg_catalog.jsonb_array_elements(v_document -> 'sections') AS section
        GROUP BY section.value ->> 'id' HAVING pg_catalog.count(*) > 1
    ) OR EXISTS (
        SELECT section.value ->> 'order'
        FROM pg_catalog.jsonb_array_elements(v_document -> 'sections') AS section
        GROUP BY section.value ->> 'order' HAVING pg_catalog.count(*) > 1
    ) OR EXISTS (
        SELECT field.value ->> 'field_id'
        FROM pg_catalog.jsonb_array_elements(v_document -> 'sections') AS section
        CROSS JOIN pg_catalog.jsonb_array_elements(
            section.value -> 'fields'
        ) AS field
        GROUP BY field.value ->> 'field_id' HAVING pg_catalog.count(*) > 1
    ) OR EXISTS (
        SELECT section.value ->> 'id', field.value ->> 'order'
        FROM pg_catalog.jsonb_array_elements(v_document -> 'sections') AS section
        CROSS JOIN pg_catalog.jsonb_array_elements(
            section.value -> 'fields'
        ) AS field
        GROUP BY section.value ->> 'id', field.value ->> 'order'
        HAVING pg_catalog.count(*) > 1
    ) THEN
        RETURN false;
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_definition_valid_v1(text)
    FROM PUBLIC;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_translations_valid_v2(
    p_translations jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF pg_catalog.jsonb_typeof(p_translations) IS DISTINCT FROM 'array' THEN
        RETURN false;
    END IF;
    IF pg_catalog.jsonb_array_length(p_translations) > 32 THEN
        RETURN false;
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(p_translations) AS item(value)
        WHERE pg_catalog.jsonb_typeof(item.value) IS DISTINCT FROM 'object'
           OR (SELECT pg_catalog.count(*)
               FROM pg_catalog.jsonb_object_keys(item.value)) <> 3
           OR EXISTS (
                SELECT 1
                FROM pg_catalog.jsonb_object_keys(item.value) AS present(key)
                WHERE present.key NOT IN ('help', 'label', 'locale')
           )
           OR pg_catalog.jsonb_typeof(item.value -> 'help') IS DISTINCT FROM 'string'
           OR pg_catalog.jsonb_typeof(item.value -> 'label') IS DISTINCT FROM 'string'
           OR pg_catalog.jsonb_typeof(item.value -> 'locale') IS DISTINCT FROM 'string'
           OR item.value ->> 'locale' !~
              '^[a-z]{2,3}(-[A-Z][a-z]{3})?(-([A-Z]{2}|[0-9]{3}))?$'
           OR pg_catalog.char_length(item.value ->> 'label') NOT BETWEEN 1 AND 200
           OR pg_catalog.char_length(item.value ->> 'help') > 1000
           OR item.value ->> 'label' ~ '[[:cntrl:]]'
           OR item.value ->> 'help' ~ '[[:cntrl:]]'
    ) OR EXISTS (
        SELECT item.value ->> 'locale'
        FROM pg_catalog.jsonb_array_elements(p_translations) AS item(value)
        GROUP BY item.value ->> 'locale'
        HAVING pg_catalog.count(*) > 1
    ) OR p_translations IS DISTINCT FROM COALESCE(
        (SELECT pg_catalog.jsonb_agg(item.value ORDER BY item.value ->> 'locale')
         FROM pg_catalog.jsonb_array_elements(p_translations) AS item(value)),
        '[]'::jsonb
    ) THEN
        RETURN false;
    END IF;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_translations_valid_v2(jsonb)
    FROM PUBLIC;

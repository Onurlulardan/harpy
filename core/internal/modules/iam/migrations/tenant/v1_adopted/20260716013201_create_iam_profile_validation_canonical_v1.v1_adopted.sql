-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_validation_canonical_v1(p_validation jsonb)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF pg_catalog.jsonb_typeof(p_validation) IS DISTINCT FROM 'object' OR
       NOT (p_validation ? 'maximum_bytes') OR EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_object_keys(p_validation) AS present(key)
        WHERE present.key NOT IN (
            'decimal_precision', 'decimal_scale', 'enum_values',
            'integer_maximum', 'integer_minimum', 'maximum_bytes',
            'maximum_items'
        )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_profile_validation_contract_guard',
            MESSAGE = 'IAM profile validation contract is invalid';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_each(p_validation) AS entry(key, value)
        WHERE (entry.key IN (
                'decimal_precision', 'decimal_scale',
                'maximum_bytes', 'maximum_items'
              ) AND pg_catalog.jsonb_typeof(entry.value) <> 'number')
           OR (entry.key IN ('integer_maximum', 'integer_minimum') AND
                pg_catalog.jsonb_typeof(entry.value) <> 'string')
           OR (entry.key = 'enum_values' AND (
                pg_catalog.jsonb_typeof(entry.value) <> 'array' OR EXISTS (
                    SELECT 1
                    FROM pg_catalog.jsonb_array_elements(entry.value)
                        AS member(value)
                    WHERE pg_catalog.jsonb_typeof(member.value) <> 'string'
                )
           ))
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'iam_profile_validation_contract_guard',
            MESSAGE = 'IAM profile validation value type is invalid';
    END IF;
    RETURN (
        SELECT '{' || pg_catalog.string_agg(
            pg_catalog.to_jsonb(entry.key)::text || ':' || CASE
                WHEN entry.key = 'enum_values' THEN (
                    SELECT '[' || COALESCE(pg_catalog.string_agg(
                        member.value::text, ',' ORDER BY member.ordinal
                    ), '') || ']'
                    FROM pg_catalog.jsonb_array_elements(entry.value)
                        WITH ORDINALITY AS member(value, ordinal)
                )
                ELSE entry.value::text
            END,
            ',' ORDER BY entry.key
        ) || '}'
        FROM pg_catalog.jsonb_each(p_validation) AS entry(key, value)
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_validation_canonical_v1(jsonb) FROM PUBLIC;

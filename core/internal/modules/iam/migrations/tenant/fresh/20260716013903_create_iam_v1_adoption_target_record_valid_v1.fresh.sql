-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_v1_adoption_target_record_valid_v1(
    p_key jsonb,
    p_value jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
BEGIN
    RETURN
        pg_catalog.jsonb_typeof(p_value) = 'object' AND
        (SELECT pg_catalog.array_agg(key ORDER BY key)
         FROM pg_catalog.jsonb_object_keys(p_value) AS key)
        = ARRAY[
            'source_category','source_key','target_key','target_record','target_table'
        ]::text[] AND
        p_value - 'target_record' = p_key AND
        pg_catalog.jsonb_typeof(p_value->'target_record') = 'object' AND
        p_value->'target_record' <> '{}'::jsonb AND
        pg_catalog.octet_length((p_value->'target_record')::text) <= 1048576;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_v1_adoption_target_record_valid_v1(
    jsonb, jsonb
) FROM PUBLIC;

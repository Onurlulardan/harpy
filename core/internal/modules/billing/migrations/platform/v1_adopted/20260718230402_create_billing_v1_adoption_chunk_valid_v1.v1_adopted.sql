-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.billing_v1_adoption_chunk_valid_v1(
    p_key jsonb,
    p_after jsonb
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_bytes bytea;
BEGIN
    IF pg_catalog.jsonb_typeof(p_key) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(p_key) AS key)
       IS DISTINCT FROM ARRAY['chunk_ordinal']::text[] OR
       pg_catalog.jsonb_typeof(p_after) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(p_after) AS key)
       IS DISTINCT FROM ARRAY[
           'chunk_base64', 'chunk_count', 'chunk_ordinal', 'chunk_sha256'
       ]::text[] OR
       (p_after->>'chunk_ordinal')::integer IS DISTINCT FROM
           (p_key->>'chunk_ordinal')::integer OR
       (p_after->>'chunk_count')::integer NOT BETWEEN 1 AND 65536 OR
       (p_after->>'chunk_ordinal')::integer NOT BETWEEN 1 AND
           (p_after->>'chunk_count')::integer OR
       p_after->>'chunk_sha256' !~ '^[0-9a-f]{64}$' THEN
        RETURN false;
    END IF;
    v_bytes := pg_catalog.decode(p_after->>'chunk_base64', 'base64');
    RETURN pg_catalog.octet_length(v_bytes) BETWEEN 1 AND 6291456 AND
        pg_catalog.encode(pg_catalog.sha256(v_bytes), 'hex') =
            p_after->>'chunk_sha256';
EXCEPTION WHEN OTHERS THEN
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.billing_v1_adoption_chunk_valid_v1(jsonb, jsonb)
    FROM PUBLIC;

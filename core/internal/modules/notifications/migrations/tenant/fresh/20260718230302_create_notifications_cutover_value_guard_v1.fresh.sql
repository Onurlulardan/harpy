-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION notifications_v1_adoption_value_valid_v1(p_key jsonb,p_value jsonb)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_target_id uuid;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       pg_catalog.jsonb_typeof(p_key) IS DISTINCT FROM 'object' OR
       pg_catalog.jsonb_typeof(p_value) IS DISTINCT FROM 'object' THEN
        RETURN false;
    END IF;
    IF (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(p_key) AS key) IS DISTINCT FROM ARRAY[
           'source_identity','source_kind','target_identity','target_kind'
       ]::text[] OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(p_value) AS key) IS DISTINCT FROM ARRAY[
           'source_identity','source_kind','source_sha256','target_id',
           'target_identity','target_kind','target_record'
       ]::text[] OR
       p_value->>'source_identity' IS DISTINCT FROM p_key->>'source_identity' OR
       p_value->>'source_kind' IS DISTINCT FROM p_key->>'source_kind' OR
       p_value->>'target_identity' IS DISTINCT FROM p_key->>'target_identity' OR
       p_value->>'target_kind' IS DISTINCT FROM p_key->>'target_kind' OR
       p_value->>'source_kind' IS DISTINCT FROM 'notification' OR
       p_value->>'target_kind' NOT IN ('message','effect','delivery','attempt','receipt') OR
       pg_catalog.octet_length(p_value->>'source_identity') NOT BETWEEN 1 AND 512 OR
       p_value->>'source_identity' OPERATOR(pg_catalog.~) '[[:cntrl:]]' OR
       pg_catalog.octet_length(p_value->>'target_identity') NOT BETWEEN 1 AND 1024 OR
       p_value->>'target_identity' OPERATOR(pg_catalog.~) '[[:cntrl:]]' OR
       p_value->>'source_sha256' !~ '^[0-9a-f]{64}$' THEN
        RETURN false;
    END IF;
    BEGIN
        v_target_id := (p_value->>'target_id')::uuid;
    EXCEPTION WHEN invalid_text_representation THEN
        RETURN false;
    END;
    RETURN notifications_v1_adoption_target_record_valid_v1(
        p_value->>'target_kind',p_value->>'target_identity',
        v_target_id,p_value->'target_record'
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION notifications_v1_adoption_value_valid_v1(
    jsonb,jsonb
) FROM PUBLIC;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_v1_adoption_value_valid_v1(
    p_key jsonb,p_value jsonb
) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_disposition text;
    v_source jsonb;
    v_target jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_adopter' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       pg_catalog.jsonb_typeof(p_key) IS DISTINCT FROM 'object' OR
       pg_catalog.jsonb_typeof(p_value) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(p_key) AS key)
         IS DISTINCT FROM ARRAY['legacy_endpoint_id']::text[] OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(p_value) AS key)
         IS DISTINCT FROM ARRAY[
          'disposition','reason_code','source_document','source_sha256',
          'target_document','target_sha256'
         ]::text[] THEN
        RETURN false;
    END IF;
    v_disposition:=p_value->>'disposition';
    v_source:=p_value->'source_document';
    v_target:=p_value->'target_document';
    IF pg_catalog.jsonb_typeof(v_source) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_source) AS key)
         IS DISTINCT FROM ARRAY[
          'config_json','created_at','created_by','deleted_at','description',
          'id','is_active','is_public','method','name','slug','tenant_id',
          'updated_at','updated_by'
         ]::text[] OR
       (p_key->>'legacy_endpoint_id')::uuid IS DISTINCT FROM
         (v_source->>'id')::uuid OR
       p_value->>'source_sha256' !~ '^[0-9a-f]{64}$' OR
       pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
         v_source::text,'UTF8'
       )),'hex') IS DISTINCT FROM p_value->>'source_sha256' OR
       pg_catalog.jsonb_typeof(v_source->'config_json') IS DISTINCT FROM 'object' OR
       v_source->>'method' NOT IN ('GET','POST') OR
       pg_catalog.octet_length(v_source->>'slug') NOT BETWEEN 1 AND 255 OR
       pg_catalog.octet_length(v_source->>'name') NOT BETWEEN 1 AND 255 THEN
        RETURN false;
    END IF;
    IF v_disposition IN ('blocked','retired') THEN
        RETURN p_value->'target_document'='null'::jsonb AND
          p_value->'target_sha256'='null'::jsonb AND
          p_value->>'reason_code'~'^[A-Z][A-Z0-9_]{0,63}$';
    END IF;
    IF v_disposition IS DISTINCT FROM 'adopted' OR
       p_value->'reason_code'<>'null'::jsonb OR
       v_source->>'is_active'<>'true' OR v_source->'deleted_at'<>'null'::jsonb OR
       pg_catalog.jsonb_typeof(v_target) IS DISTINCT FROM 'object' OR
       (SELECT pg_catalog.array_agg(key ORDER BY key)
        FROM pg_catalog.jsonb_object_keys(v_target) AS key)
         IS DISTINCT FROM ARRAY[
          'approval','dependencies','differential','draft_version',
          'legacy_route','public_pointer','published_version','query','release'
         ]::text[] OR
       p_value->>'target_sha256' !~ '^[0-9a-f]{64}$' OR
       pg_catalog.encode(pg_catalog.sha256(pg_catalog.convert_to(
         v_target::text,'UTF8'
       )),'hex') IS DISTINCT FROM p_value->>'target_sha256' OR
       pg_catalog.jsonb_typeof(v_target->'query') IS DISTINCT FROM 'object' OR
       pg_catalog.jsonb_typeof(v_target->'published_version') IS DISTINCT FROM 'object' OR
       pg_catalog.jsonb_typeof(v_target->'draft_version') IS DISTINCT FROM 'object' OR
       pg_catalog.jsonb_typeof(v_target->'dependencies') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(v_target->'dependencies') NOT BETWEEN 1 AND 32 OR
       pg_catalog.jsonb_typeof(v_target->'differential') IS DISTINCT FROM 'object' THEN
        RETURN false;
    END IF;
    RETURN (v_source->>'is_public'='true' AND
            pg_catalog.jsonb_typeof(v_target->'release')='object' AND
            pg_catalog.jsonb_typeof(v_target->'approval')='object' AND
            pg_catalog.jsonb_typeof(v_target->'public_pointer')='object' AND
            pg_catalog.jsonb_typeof(v_target->'legacy_route')='object')
        OR (v_source->>'is_public'='false' AND
            v_target->'release'='null'::jsonb AND
            v_target->'approval'='null'::jsonb AND
            v_target->'public_pointer'='null'::jsonb AND
            v_target->'legacy_route'='null'::jsonb);
EXCEPTION WHEN invalid_text_representation OR numeric_value_out_of_range THEN
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_v1_adoption_value_valid_v1(jsonb,jsonb)
FROM PUBLIC;

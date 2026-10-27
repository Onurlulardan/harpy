-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_semantic_sha256_v1(
    p_field_id uuid,
    p_resource_kind text,
    p_field_key text,
    p_value_kind text,
    p_cardinality text,
    p_required boolean,
    p_classification text,
    p_normalization_profile text
)
RETURNS bytea
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_document text;
BEGIN
    v_document := '{"cardinality":' ||
        pg_catalog.to_jsonb(p_cardinality)::text ||
        ',"classification":' || pg_catalog.to_jsonb(p_classification)::text ||
        ',"contract":"forgecrud/iam-profile-field-semantic/v1"' ||
        ',"id":' || pg_catalog.to_jsonb(p_field_id::text)::text ||
        ',"key":' || pg_catalog.to_jsonb(p_field_key)::text ||
        ',"logical_type":' || pg_catalog.to_jsonb(p_value_kind)::text ||
        ',"normalization_profile":' ||
        pg_catalog.to_jsonb(p_normalization_profile)::text ||
        ',"required":' || pg_catalog.to_jsonb(p_required)::text ||
        ',"resource_kind":' || pg_catalog.to_jsonb(p_resource_kind)::text || '}';
    RETURN pg_catalog.sha256(pg_catalog.convert_to(v_document, 'UTF8'));
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_semantic_sha256_v1(
    uuid,text,text,text,text,boolean,text,text
) FROM PUBLIC;

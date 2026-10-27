-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_backfill_mapping_create_v1(
    p_plan_id uuid, p_resource_kind text, p_source_schema_id uuid,
    p_target_schema_id uuid, p_mappings jsonb
)
RETURNS integer
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_count integer;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR p_plan_id IS NULL OR
       p_resource_kind NOT IN ('user','organization','role') OR
       p_source_schema_id IS NULL OR p_target_schema_id IS NULL OR
       pg_catalog.jsonb_typeof(p_mappings)<>'array' OR
       pg_catalog.jsonb_array_length(p_mappings) NOT BETWEEN 1 AND 500 OR EXISTS (
           SELECT 1 FROM pg_catalog.jsonb_array_elements(p_mappings)
                WITH ORDINALITY AS mapping(value,ordinal)
           WHERE pg_catalog.jsonb_typeof(mapping.value)<>'object'
              OR (mapping.value->>'ordinal')::integer<>mapping.ordinal
              OR mapping.value->>'converter_kind' NOT IN ('copy_v1','copy_normalize_v1')
              OR pg_catalog.octet_length(pg_catalog.decode(
                    mapping.value->>'semantic_sha256','hex'))<>32
              OR EXISTS (
                   SELECT 1 FROM pg_catalog.jsonb_object_keys(mapping.value) AS key
                   WHERE key NOT IN ('ordinal','source_field_id','target_field_id',
                                     'converter_kind','semantic_sha256')
              )
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM profile backfill mapping input is invalid';
    END IF;
    INSERT INTO iam_profile_backfill_mappings (
        plan_id,ordinal,source_field_id,target_field_id,converter_kind,semantic_sha256
    )
    SELECT p_plan_id,mapping.ordinal,(mapping.value->>'source_field_id')::uuid,
        (mapping.value->>'target_field_id')::uuid,mapping.value->>'converter_kind',
        pg_catalog.decode(mapping.value->>'semantic_sha256','hex')
    FROM pg_catalog.jsonb_array_elements(p_mappings)
         WITH ORDINALITY AS mapping(value,ordinal);
    GET DIAGNOSTICS v_count=ROW_COUNT;
    IF EXISTS (
        SELECT 1 FROM iam_profile_backfill_mappings AS mapping
        LEFT JOIN iam_profile_field_revisions AS source
          ON source.schema_version_id=p_source_schema_id
         AND source.resource_kind=p_resource_kind
         AND source.field_id=mapping.source_field_id
        LEFT JOIN iam_profile_field_revisions AS target
          ON target.schema_version_id=p_target_schema_id
         AND target.resource_kind=p_resource_kind
         AND target.field_id=mapping.target_field_id
        WHERE mapping.plan_id=p_plan_id AND (
            source.field_id IS NULL OR target.field_id IS NULL OR
            target.semantic_sha256<>mapping.semantic_sha256 OR
            (mapping.converter_kind='copy_v1' AND
             source.semantic_sha256<>target.semantic_sha256)
        )
    ) OR EXISTS (
        SELECT 1 FROM iam_profile_field_revisions AS required
        WHERE required.schema_version_id=p_target_schema_id
          AND required.resource_kind=p_resource_kind AND required.field_state='active'
          AND required.required AND NOT EXISTS (
              SELECT 1 FROM iam_profile_backfill_mappings AS mapping
              WHERE mapping.plan_id=p_plan_id AND mapping.target_field_id=required.field_id
          )
    ) THEN
        RAISE EXCEPTION USING ERRCODE='23514',
            MESSAGE='IAM profile backfill mapping is inconsistent';
    END IF;
    RETURN v_count;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_backfill_mapping_create_v1(
    uuid,text,uuid,uuid,jsonb
) FROM PUBLIC;


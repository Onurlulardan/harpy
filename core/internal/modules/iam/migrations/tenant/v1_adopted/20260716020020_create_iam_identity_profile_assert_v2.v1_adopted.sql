-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_identity_profile_assert_v2(
    p_resource_kind text, p_schema_id uuid,
    p_profile_document bytea, p_profile_sha256 bytea
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_values jsonb;
    v_source_schema_id uuid;
    v_plan_id uuid;
    v_target_schema_id uuid;
BEGIN
    IF p_resource_kind NOT IN ('user', 'organization', 'role') OR
       p_schema_id IS NULL OR
       pg_catalog.octet_length(p_profile_document) > 1048576 OR
       pg_catalog.octet_length(p_profile_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.sha256(p_profile_document) IS DISTINCT FROM p_profile_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM identity profile input is invalid';
    END IF;
    SELECT source.id INTO v_source_schema_id
    FROM iam_profile_schema_versions AS source
    WHERE source.resource_kind = p_resource_kind AND source.state = 'active'
    FOR SHARE;
    SELECT plan.id, plan.target_schema_id INTO v_plan_id, v_target_schema_id
    FROM iam_profile_backfill_plans AS plan
    JOIN iam_profile_schema_versions AS source
      ON source.id = plan.source_schema_id AND source.state = 'active'
    JOIN iam_profile_schema_versions AS target
      ON target.id = plan.target_schema_id AND target.state = 'draft'
    WHERE plan.resource_kind = p_resource_kind
      AND plan.source_schema_id = v_source_schema_id
      AND plan.state IN ('planned', 'running', 'verified')
    FOR SHARE OF plan, source, target;
    IF v_source_schema_id IS NULL OR
       (v_plan_id IS NULL AND p_schema_id <> v_source_schema_id) OR
       (v_plan_id IS NOT NULL AND p_schema_id <> v_target_schema_id) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM identity profile write schema is stale';
    END IF;
    v_values := pg_catalog.convert_from(p_profile_document, 'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_values) <> 'object' OR EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_each(v_values) AS entry(key, value)
        WHERE entry.key !~
            '^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
           OR pg_catalog.jsonb_typeof(entry.value) = 'null'
           OR NOT EXISTS (
              SELECT 1 FROM iam_profile_field_revisions AS field
              WHERE field.schema_version_id = p_schema_id
                AND field.resource_kind = p_resource_kind
                AND field.field_id = entry.key::uuid AND (
                    field.field_state = 'active' OR
                    field.field_state = 'retired' AND v_plan_id IS NOT NULL AND EXISTS (
                        SELECT 1 FROM iam_profile_field_revisions AS source_field
                        WHERE source_field.schema_version_id = v_source_schema_id
                          AND source_field.resource_kind = p_resource_kind
                          AND source_field.field_id = field.field_id
                          AND source_field.semantic_sha256 = field.semantic_sha256
                    )
                )
           )
    ) OR EXISTS (
        SELECT 1 FROM iam_profile_field_revisions AS required
        WHERE required.schema_version_id = p_schema_id
          AND required.resource_kind = p_resource_kind
          AND required.field_state = 'active' AND required.required
          AND NOT v_values ? required.field_id::text
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM identity profile contract is invalid';
    END IF;
    RETURN v_values;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_identity_profile_assert_v2(
    text,uuid,bytea,bytea
) FROM PUBLIC;

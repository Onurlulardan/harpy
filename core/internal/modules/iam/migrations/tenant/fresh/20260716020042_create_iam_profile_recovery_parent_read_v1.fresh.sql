-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_recovery_parent_read_v1(
    p_scope_schema name, p_parent_plan_id uuid, p_actor_id uuid,
    p_manage_permission_id uuid, p_expected_token_version bigint,
    p_expected_authorization_version bigint
)
RETURNS TABLE (
    plan_id uuid, recovery_root_plan_id uuid, recovery_sequence bigint,
    plan_sha256 bytea, resource_kind text, source_schema_id uuid,
    target_schema_id uuid, converter_version text, existing_child_plan_id uuid,
    source_snapshot jsonb, target_snapshot jsonb, mappings jsonb
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_root_plan_id uuid;
    v_sequence bigint;
    v_plan_sha256 bytea;
    v_resource_kind text;
    v_source_schema_id uuid;
    v_target_schema_id uuid;
    v_converter_version text;
    v_child_plan_id uuid;
    v_source_snapshot jsonb;
    v_target_snapshot jsonb;
    v_mappings jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_parent_plan_id IS NULL OR NOT iam_profile_actor_assert_v1(
           p_scope_schema, p_actor_id, p_manage_permission_id,
           p_expected_token_version, p_expected_authorization_version
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile recovery parent authority is denied';
    END IF;
    SELECT parent.recovery_root_plan_id, COALESCE(parent.recovery_sequence, 0),
           parent.plan_sha256, parent.resource_kind, parent.source_schema_id,
           parent.target_schema_id, parent.converter_version
    INTO v_root_plan_id, v_sequence, v_plan_sha256, v_resource_kind,
         v_source_schema_id, v_target_schema_id, v_converter_version
    FROM iam_profile_backfill_plans AS parent
    WHERE parent.id = p_parent_plan_id AND parent.state = 'fenced'
    FOR UPDATE OF parent;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile recovery parent is unavailable';
    END IF;
    SELECT child.id INTO v_child_plan_id
    FROM iam_profile_backfill_plans AS child
    WHERE child.recovery_of_plan_id = p_parent_plan_id;
    IF v_child_plan_id IS NOT NULL THEN
        RETURN QUERY SELECT p_parent_plan_id, v_root_plan_id, v_sequence,
            v_plan_sha256, v_resource_kind, v_source_schema_id,
            v_target_schema_id, v_converter_version, v_child_plan_id,
            NULL::jsonb, NULL::jsonb, NULL::jsonb;
        RETURN;
    END IF;
    PERFORM 1
    FROM iam_profile_schema_versions AS source
    JOIN iam_profile_schema_versions AS target
      ON target.id = v_target_schema_id
     AND target.resource_kind = v_resource_kind AND target.state = 'draft'
     AND target.version = source.version + 1
    WHERE source.id = v_source_schema_id
      AND source.resource_kind = v_resource_kind AND source.state = 'active'
    FOR SHARE OF source, target;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile recovery lineage is unavailable';
    END IF;
    v_source_snapshot := iam_profile_schema_snapshot_json_v1(v_source_schema_id);
    v_target_snapshot := iam_profile_schema_snapshot_json_v1(v_target_schema_id);
    SELECT COALESCE(pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'source_field_id', mapping.source_field_id,
        'target_field_id', mapping.target_field_id,
        'converter_kind', mapping.converter_kind,
        'semantic_sha256', pg_catalog.encode(mapping.semantic_sha256, 'hex')
    ) ORDER BY mapping.ordinal), '[]'::jsonb)
    INTO v_mappings
    FROM iam_profile_backfill_mappings AS mapping
    WHERE mapping.plan_id = p_parent_plan_id;
    IF v_source_snapshot IS NULL OR v_target_snapshot IS NULL OR
       pg_catalog.jsonb_array_length(v_mappings) < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile recovery parent contract is unavailable';
    END IF;
    RETURN QUERY SELECT p_parent_plan_id, v_root_plan_id, v_sequence,
        v_plan_sha256, v_resource_kind, v_source_schema_id, v_target_schema_id,
        v_converter_version, v_child_plan_id, v_source_snapshot,
        v_target_snapshot, v_mappings;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_recovery_parent_read_v1(
    name,uuid,uuid,uuid,bigint,bigint
) FROM PUBLIC;

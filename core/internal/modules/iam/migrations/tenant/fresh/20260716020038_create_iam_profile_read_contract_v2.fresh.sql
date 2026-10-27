-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_read_contract_v2(
    p_scope_schema name, p_resource_kind text
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_source_schema_id uuid;
    v_plan_id uuid;
    v_target_schema_id uuid;
    v_source jsonb;
    v_target jsonb;
    v_mappings jsonb;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_resource_kind NOT IN ('user', 'organization', 'role') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'IAM profile read contract scope is invalid';
    END IF;
    SELECT source.id INTO v_source_schema_id
    FROM iam_profile_schema_versions AS source
    WHERE source.resource_kind = p_resource_kind AND source.state = 'active';
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM active profile read schema is unavailable';
    END IF;
    SELECT plan.id, plan.target_schema_id
    INTO v_plan_id, v_target_schema_id
    FROM iam_profile_backfill_plans AS plan
    JOIN iam_profile_schema_versions AS source
      ON source.id = plan.source_schema_id
     AND source.resource_kind = plan.resource_kind AND source.state = 'active'
    JOIN iam_profile_schema_versions AS target
      ON target.id = plan.target_schema_id
     AND target.resource_kind = plan.resource_kind AND target.state = 'draft'
    WHERE plan.resource_kind = p_resource_kind
      AND plan.source_schema_id = v_source_schema_id
      AND plan.state IN ('planned', 'running', 'verified', 'fenced')
    ORDER BY (plan.state IN ('planned', 'running', 'verified')) DESC,
             COALESCE(plan.recovery_sequence, 0) DESC, plan.created_at DESC
    LIMIT 1;
    v_source := iam_profile_schema_snapshot_json_v1(v_source_schema_id);
    IF v_source IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM source profile read contract is unavailable';
    END IF;
    IF v_plan_id IS NOT NULL THEN
        v_target := iam_profile_schema_snapshot_json_v1(v_target_schema_id);
        SELECT COALESCE(pg_catalog.jsonb_agg(
            pg_catalog.jsonb_build_object(
                'source_field_id', mapping.source_field_id,
                'target_field_id', mapping.target_field_id,
                'converter_kind', mapping.converter_kind,
                'semantic_sha256', pg_catalog.encode(
                    mapping.semantic_sha256, 'hex'
                )
            ) ORDER BY mapping.ordinal
        ), '[]'::jsonb)
        INTO v_mappings
        FROM iam_profile_backfill_mappings AS mapping
        WHERE mapping.plan_id = v_plan_id;
        IF v_target IS NULL OR pg_catalog.jsonb_array_length(v_mappings) < 1 THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'IAM target profile read contract is unavailable';
        END IF;
    END IF;
    RETURN pg_catalog.jsonb_build_object(
        'contract', 'forgecrud/iam-profile-write-contract/v1',
        'plan_id', v_plan_id,
        'source', v_source,
        'target', v_target,
        'mappings', COALESCE(v_mappings, '[]'::jsonb)
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_read_contract_v2(name,text)
    FROM PUBLIC;

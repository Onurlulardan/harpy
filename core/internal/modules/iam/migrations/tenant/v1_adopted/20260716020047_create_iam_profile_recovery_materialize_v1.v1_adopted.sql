-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_recovery_materialize_v1(
    p_tenant_id uuid, p_plan_document bytea, p_plan_sha256 bytea,
    p_created_at timestamptz
)
RETURNS TABLE (
    plan_id uuid, resource_kind text, source_schema_id uuid,
    target_schema_id uuid, remaining_source_count bigint
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_plan jsonb;
    v_plan_id uuid;
    v_parent_id uuid;
    v_root_id uuid;
    v_sequence bigint;
    v_resource text;
    v_source_id uuid;
    v_target_id uuid;
    v_target_version bigint;
    v_converter text;
    v_parent_sha bytea;
    v_target_definition_sha bytea;
    v_inherited_count bigint;
    v_inherited_sha bytea;
    v_remaining_count bigint;
    v_remaining_sha bytea;
    v_expected_target_sha bytea;
    v_initial_rolling bytea;
    v_page_size integer;
    v_page_count integer;
    v_parent_mappings jsonb;
    v_expected_root uuid;
    v_expected_sequence bigint;
    v_target_count bigint;
    v_actual_target_sha bytea;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR p_tenant_id IS NULL OR
       pg_catalog.octet_length(p_plan_sha256) <> 32 OR
       pg_catalog.octet_length(p_plan_document) NOT BETWEEN 2 AND 16777216 OR
       pg_catalog.sha256(p_plan_document) <> p_plan_sha256 OR
       p_created_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile recovery materialization input is invalid';
    END IF;
    v_plan := pg_catalog.convert_from(p_plan_document,'UTF8')::jsonb;
    IF pg_catalog.jsonb_typeof(v_plan) <> 'object' OR EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_object_keys(v_plan) AS key
        WHERE key NOT IN (
            'id','resource_kind','recovery_of_plan_id','recovery_root_plan_id',
            'recovery_sequence','parent_plan_sha256','source_schema_id',
            'target_schema_id','target_version','target_definition_sha256',
            'converter_version','inherited_target_count','inherited_target_sha256',
            'remaining_source_count','remaining_source_sha256',
            'expected_remaining_target_sha256','page_size',
            'initial_rolling_sha256','mappings','pages'
        )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile recovery plan document is invalid';
    END IF;
    v_plan_id := (v_plan->>'id')::uuid;
    v_parent_id := (v_plan->>'recovery_of_plan_id')::uuid;
    v_root_id := (v_plan->>'recovery_root_plan_id')::uuid;
    v_sequence := (v_plan->>'recovery_sequence')::bigint;
    v_resource := v_plan->>'resource_kind';
    v_source_id := (v_plan->>'source_schema_id')::uuid;
    v_target_id := (v_plan->>'target_schema_id')::uuid;
    v_target_version := (v_plan->>'target_version')::bigint;
    v_converter := v_plan->>'converter_version';
    v_parent_sha := pg_catalog.decode(v_plan->>'parent_plan_sha256','hex');
    v_target_definition_sha := pg_catalog.decode(
        v_plan->>'target_definition_sha256','hex');
    v_inherited_count := (v_plan->>'inherited_target_count')::bigint;
    v_inherited_sha := pg_catalog.decode(
        v_plan->>'inherited_target_sha256','hex');
    v_remaining_count := (v_plan->>'remaining_source_count')::bigint;
    v_remaining_sha := pg_catalog.decode(v_plan->>'remaining_source_sha256','hex');
    v_expected_target_sha := pg_catalog.decode(
        v_plan->>'expected_remaining_target_sha256','hex');
    v_page_size := (v_plan->>'page_size')::integer;
    v_initial_rolling := pg_catalog.decode(
        v_plan->>'initial_rolling_sha256','hex');
    v_page_count := pg_catalog.jsonb_array_length(v_plan->'pages');
    IF v_plan_id IS NULL OR v_parent_id IS NULL OR v_root_id IS NULL OR
       v_plan_id IN (v_parent_id,v_root_id) OR v_sequence < 1 OR
       v_resource NOT IN ('user','organization','role') OR v_source_id IS NULL OR
       v_target_id IS NULL OR v_source_id = v_target_id OR v_target_version < 2 OR
       v_converter <> 'typed_map_v1' OR v_inherited_count < 0 OR
       v_remaining_count < 0 OR v_page_size NOT BETWEEN 1 AND 250 OR
       pg_catalog.jsonb_typeof(v_plan->'mappings') <> 'array' OR
       pg_catalog.jsonb_typeof(v_plan->'pages') <> 'array' OR
       pg_catalog.octet_length(v_parent_sha) <> 32 OR
       pg_catalog.octet_length(v_target_definition_sha) <> 32 OR
       pg_catalog.octet_length(v_inherited_sha) <> 32 OR
       pg_catalog.octet_length(v_remaining_sha) <> 32 OR
       pg_catalog.octet_length(v_expected_target_sha) <> 32 OR
       pg_catalog.octet_length(v_initial_rolling) <> 32 OR
       (v_remaining_count = 0 AND v_page_count <> 0) OR
       (v_remaining_count > 0 AND v_page_count NOT BETWEEN 1 AND 65536) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile recovery plan contract is invalid';
    END IF;
    SELECT COALESCE(parent.recovery_root_plan_id,parent.id),
           COALESCE(parent.recovery_sequence,0) + 1
    INTO v_expected_root,v_expected_sequence
    FROM iam_profile_backfill_plans AS parent
    JOIN iam_profile_schema_versions AS source
      ON source.id = parent.source_schema_id AND source.state = 'active'
     AND source.resource_kind = parent.resource_kind
    JOIN iam_profile_schema_versions AS target
      ON target.id = parent.target_schema_id AND target.state = 'draft'
     AND target.resource_kind = parent.resource_kind
     AND target.version = source.version + 1
    WHERE parent.id = v_parent_id AND parent.state = 'fenced'
      AND parent.resource_kind = v_resource
      AND parent.source_schema_id = v_source_id
      AND parent.target_schema_id = v_target_id
      AND parent.converter_version = v_converter
      AND parent.plan_sha256 = v_parent_sha
      AND target.version = v_target_version
      AND target.definition_sha256 = v_target_definition_sha
      AND NOT EXISTS (SELECT 1 FROM iam_profile_backfill_plans AS child
                      WHERE child.recovery_of_plan_id = parent.id)
    FOR UPDATE OF parent FOR SHARE OF source, target;
    IF v_expected_root IS NULL OR v_root_id <> v_expected_root OR
       v_sequence <> v_expected_sequence THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile recovery parent changed';
    END IF;
    SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'ordinal', mapping.ordinal, 'source_field_id', mapping.source_field_id,
        'target_field_id', mapping.target_field_id,
        'converter_kind', mapping.converter_kind,
        'semantic_sha256', pg_catalog.encode(mapping.semantic_sha256,'hex')
    ) ORDER BY mapping.ordinal) INTO v_parent_mappings
    FROM iam_profile_backfill_mappings AS mapping WHERE mapping.plan_id = v_parent_id;
    IF v_parent_mappings IS NULL OR v_plan->'mappings' <> v_parent_mappings THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile recovery mapping contract changed';
    END IF;
    INSERT INTO iam_profile_backfill_plans (
        id,resource_kind,source_schema_id,target_schema_id,converter_version,
        plan_sha256,source_count,expected_source_sha256,expected_target_sha256,
        page_size,page_count,state,revision,created_at,updated_at,
        recovery_of_plan_id,recovery_root_plan_id,recovery_sequence,
        parent_plan_sha256,target_definition_sha256,inherited_target_count,
        inherited_target_sha256,remaining_source_count,remaining_source_sha256,
        expected_remaining_target_sha256,initial_rolling_sha256
    ) VALUES (
        v_plan_id,v_resource,v_source_id,v_target_id,v_converter,p_plan_sha256,
        v_remaining_count,v_remaining_sha,v_expected_target_sha,v_page_size,
        v_page_count,'planned',1,p_created_at,p_created_at,v_parent_id,v_root_id,
        v_sequence,v_parent_sha,v_target_definition_sha,v_inherited_count,
        v_inherited_sha,v_remaining_count,v_remaining_sha,v_expected_target_sha,
        v_initial_rolling
    );
    INSERT INTO iam_profile_backfill_mappings
    SELECT v_plan_id,mapping.ordinal,mapping.source_field_id,mapping.target_field_id,
           mapping.converter_kind,mapping.semantic_sha256 FROM iam_profile_backfill_mappings AS mapping
    WHERE mapping.plan_id = v_parent_id ORDER BY mapping.ordinal;
    IF v_remaining_count > 0 THEN
        IF iam_profile_recovery_pages_create_v1(
            p_tenant_id,v_plan_id,v_resource,v_source_id,v_target_id,
            v_remaining_count,v_page_size,v_remaining_sha,v_expected_target_sha,
            v_initial_rolling,v_inherited_count,v_inherited_sha,
            v_plan->'pages',p_created_at
        ) <> v_page_count THEN
            RAISE EXCEPTION USING ERRCODE = '23514',
                MESSAGE = 'IAM profile recovery pages are incomplete';
        END IF;
    ELSE
        SELECT evidence.target_count,evidence.target_sha256
        INTO v_target_count,v_actual_target_sha
        FROM iam_profile_recovery_target_evidence_v1(v_resource,v_target_id)
             AS evidence;
        IF v_remaining_sha <> pg_catalog.sha256(''::bytea) OR
           v_expected_target_sha <> v_initial_rolling OR
           v_target_count <> v_inherited_count OR
           v_actual_target_sha <> v_inherited_sha THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM empty profile recovery evidence is invalid';
        END IF;
    END IF;
    INSERT INTO iam_profile_backfill_checkpoints (
        plan_id,completed_pages,converted_count,error_count,rolling_sha256,
        last_page_number,revision,updated_at
    ) VALUES (v_plan_id,0,0,0,v_initial_rolling,NULL,1,p_created_at);
    RETURN QUERY SELECT v_plan_id,v_resource,v_source_id,v_target_id,
        v_remaining_count;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_profile_recovery_materialize_v1(
    uuid,bytea,bytea,timestamptz
) FROM PUBLIC;

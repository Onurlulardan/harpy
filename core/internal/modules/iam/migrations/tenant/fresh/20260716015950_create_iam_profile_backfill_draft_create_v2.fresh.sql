-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_backfill_draft_create_v2(
    p_scope_schema name, p_schema_id uuid, p_resource_kind text,
    p_version bigint, p_definition jsonb, p_definition_sha256 bytea,
    p_expected_active_schema_id uuid
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz := pg_catalog.transaction_timestamp();
    v_active_id uuid;
    v_active_version bigint;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_schema_id IS NULL OR p_expected_active_schema_id IS NULL OR
       p_schema_id IS NOT DISTINCT FROM p_expected_active_schema_id OR
       p_resource_kind NOT IN ('user', 'organization', 'role') OR
       p_version < 2 OR pg_catalog.jsonb_typeof(p_definition) <> 'object' OR
       pg_catalog.octet_length(p_definition_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_definition::text) > 4194304 OR
       p_definition ->> 'contract' IS DISTINCT FROM 'harpy/iam-profile-schema/v1' OR
       p_definition ->> 'id' IS DISTINCT FROM p_schema_id::text OR
       p_definition ->> 'resource_kind' IS DISTINCT FROM p_resource_kind OR
       p_definition ->> 'version' IS DISTINCT FROM p_version::text OR
       pg_catalog.jsonb_typeof(p_definition -> 'fields') <> 'array' OR
       pg_catalog.jsonb_array_length(p_definition -> 'fields') > 500 OR EXISTS (
           SELECT 1 FROM pg_catalog.jsonb_object_keys(p_definition) AS key
           WHERE key NOT IN ('contract', 'fields', 'id', 'resource_kind', 'version')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile backfill draft definition is invalid';
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_array_elements(p_definition -> 'fields') AS field(value)
        WHERE pg_catalog.jsonb_typeof(field.value) <> 'object'
           OR field.value ->> 'id' !~
              '^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
           OR pg_catalog.jsonb_typeof(field.value -> 'required') <> 'boolean'
           OR pg_catalog.jsonb_typeof(field.value -> 'validation') <> 'object'
           OR field.value ->> 'state' NOT IN ('active', 'retired')
           OR EXISTS (
                SELECT 1 FROM pg_catalog.jsonb_object_keys(field.value) AS key
                WHERE key NOT IN (
                    'id', 'key', 'logical_type', 'cardinality', 'required',
                    'classification', 'normalization_profile', 'state', 'validation'
                )
           )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile backfill draft field is invalid';
    END IF;
    SELECT active.id, active.version INTO v_active_id, v_active_version
    FROM iam_profile_schema_versions AS active
    WHERE active.resource_kind = p_resource_kind AND active.state = 'active'
    FOR UPDATE;
    IF NOT FOUND OR v_active_id IS DISTINCT FROM p_expected_active_schema_id OR
       p_version IS DISTINCT FROM v_active_version + 1 OR EXISTS (
           SELECT 1 FROM iam_profile_schema_versions AS draft
           WHERE draft.resource_kind = p_resource_kind AND draft.state = 'draft'
       ) OR EXISTS (
           SELECT 1 FROM iam_profile_backfill_plans AS plan
           WHERE plan.resource_kind = p_resource_kind
             AND plan.state IN ('planned', 'running', 'verified')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile backfill draft compare-and-swap failed';
    END IF;
    IF EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_array_elements(p_definition -> 'fields') AS field(value)
        LEFT JOIN iam_profile_fields AS stored
          ON stored.id = (field.value ->> 'id')::uuid
         AND stored.resource_kind = p_resource_kind
        WHERE stored.id IS NULL AND field.value ->> 'state' <> 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile backfill new field must be active';
    END IF;
    IF EXISTS (
        SELECT 1 FROM iam_profile_field_revisions AS source
        WHERE source.schema_version_id = v_active_id
          AND source.resource_kind = p_resource_kind
          AND NOT EXISTS (
              SELECT 1 FROM pg_catalog.jsonb_array_elements(
                  p_definition -> 'fields'
              ) AS target(value)
              WHERE target.value ->> 'id' = source.field_id::text
          )
    ) OR EXISTS (
        SELECT 1 FROM iam_profile_field_revisions AS source
        JOIN pg_catalog.jsonb_array_elements(
            p_definition -> 'fields'
        ) AS target(value) ON target.value ->> 'id' = source.field_id::text
        WHERE source.schema_version_id = v_active_id
          AND source.resource_kind = p_resource_kind
          AND source.field_state = 'retired'
          AND target.value ->> 'state' <> 'retired'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile backfill fields must evolve cumulatively';
    END IF;
    INSERT INTO iam_profile_fields (
        id, resource_kind, field_key, value_kind, cardinality, required,
        classification, normalization_profile, state, created_at, updated_at,
        retired_at, revision
    )
    SELECT (field.value ->> 'id')::uuid, p_resource_kind, field.value ->> 'key',
        field.value ->> 'logical_type', field.value ->> 'cardinality',
        (field.value ->> 'required')::boolean, field.value ->> 'classification',
        field.value ->> 'normalization_profile', 'active', v_now, v_now, NULL, 1
    FROM pg_catalog.jsonb_array_elements(p_definition -> 'fields') AS field(value)
    WHERE NOT EXISTS (
        SELECT 1 FROM iam_profile_fields AS stored
        WHERE stored.id = (field.value ->> 'id')::uuid
    );
    IF EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_array_elements(p_definition -> 'fields') AS field(value)
        LEFT JOIN iam_profile_fields AS stored
          ON stored.id = (field.value ->> 'id')::uuid
         AND stored.resource_kind = p_resource_kind
        WHERE stored.id IS NULL OR stored.field_key IS DISTINCT FROM field.value ->> 'key'
           OR stored.value_kind IS DISTINCT FROM field.value ->> 'logical_type'
           OR stored.cardinality IS DISTINCT FROM field.value ->> 'cardinality'
           OR stored.required IS DISTINCT FROM (field.value ->> 'required')::boolean
           OR stored.classification IS DISTINCT FROM field.value ->> 'classification'
           OR stored.normalization_profile IS DISTINCT FROM field.value ->> 'normalization_profile'
           OR (stored.state = 'retired' AND field.value ->> 'state' <> 'retired')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile backfill stable field semantic changed';
    END IF;
    INSERT INTO iam_profile_schema_versions (
        id, resource_kind, version, state, definition_sha256,
        activated_at, retired_at, created_at, updated_at, revision
    ) VALUES (
        p_schema_id, p_resource_kind, p_version, 'draft', p_definition_sha256,
        NULL, NULL, v_now, v_now, 1
    );
    INSERT INTO iam_profile_field_revisions (
        schema_version_id, resource_kind, field_id, validation_schema,
        validation_sha256, created_at, field_key, value_kind, cardinality,
        required, classification, normalization_profile, field_state,
        semantic_sha256
    )
    SELECT p_schema_id, p_resource_kind, (field.value ->> 'id')::uuid,
        field.value -> 'validation', pg_catalog.sha256(pg_catalog.convert_to(
            iam_profile_validation_canonical_v1(field.value -> 'validation'), 'UTF8'
        )), v_now, field.value ->> 'key', field.value ->> 'logical_type',
        field.value ->> 'cardinality', (field.value ->> 'required')::boolean,
        field.value ->> 'classification', field.value ->> 'normalization_profile',
        field.value ->> 'state', iam_profile_semantic_sha256_v1(
            (field.value ->> 'id')::uuid, p_resource_kind, field.value ->> 'key',
            field.value ->> 'logical_type', field.value ->> 'cardinality',
            (field.value ->> 'required')::boolean, field.value ->> 'classification',
            field.value ->> 'normalization_profile'
        )
    FROM pg_catalog.jsonb_array_elements(p_definition -> 'fields') AS field(value);
    IF iam_profile_definition_sha256_v1(p_schema_id) IS DISTINCT FROM p_definition_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile backfill draft hash is inconsistent';
    END IF;
    RETURN 1;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_backfill_draft_create_v2(
    name,uuid,text,bigint,jsonb,bytea,uuid
) FROM PUBLIC;

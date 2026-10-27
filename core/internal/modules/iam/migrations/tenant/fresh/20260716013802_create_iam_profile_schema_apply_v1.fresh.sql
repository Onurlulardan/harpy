-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_profile_schema_apply_v1(
    p_scope_schema name, p_schema_id uuid, p_resource_kind text,
    p_version bigint, p_definition jsonb, p_definition_sha256 bytea,
    p_expected_active_schema_id uuid, p_retire_field_id uuid,
    p_expected_field_revision bigint
)
RETURNS bigint
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_now timestamptz; v_active_id uuid; v_active_version bigint;
    v_definition_sha256 bytea; v_revision bigint;
BEGIN
    v_now := pg_catalog.transaction_timestamp();
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       p_scope_schema IS DISTINCT FROM pg_catalog.current_schema() OR
       p_schema_id IS NULL OR p_expected_active_schema_id IS NULL OR
       p_schema_id IS NOT DISTINCT FROM p_expected_active_schema_id OR
       p_resource_kind NOT IN ('user', 'organization', 'role') OR
       p_version < 2 OR pg_catalog.jsonb_typeof(p_definition) <> 'object' OR
       pg_catalog.octet_length(p_definition_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_definition::text) > 4194304 OR
       p_definition ->> 'contract' IS DISTINCT FROM 'forgecrud/iam-profile-schema/v1' OR
       p_definition ->> 'id' IS DISTINCT FROM p_schema_id::text OR
       p_definition ->> 'resource_kind' IS DISTINCT FROM p_resource_kind OR
       p_definition ->> 'version' IS DISTINCT FROM p_version::text OR
       pg_catalog.jsonb_typeof(p_definition -> 'fields') <> 'array' OR
       pg_catalog.jsonb_array_length(p_definition -> 'fields') > 500 OR EXISTS (
           SELECT 1 FROM pg_catalog.jsonb_object_keys(p_definition) AS key
           WHERE key NOT IN ('contract', 'fields', 'id', 'resource_kind', 'version')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile schema definition is invalid';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(p_definition -> 'fields') AS field(value)
        WHERE pg_catalog.jsonb_typeof(field.value) <> 'object'
           OR pg_catalog.jsonb_typeof(field.value -> 'id') <> 'string'
           OR field.value ->> 'id' !~
                '^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
           OR pg_catalog.jsonb_typeof(field.value -> 'key') <> 'string'
           OR pg_catalog.jsonb_typeof(field.value -> 'logical_type') <> 'string'
           OR pg_catalog.jsonb_typeof(field.value -> 'cardinality') <> 'string'
           OR pg_catalog.jsonb_typeof(field.value -> 'required') <> 'boolean'
           OR pg_catalog.jsonb_typeof(field.value -> 'classification') <> 'string'
           OR pg_catalog.jsonb_typeof(field.value -> 'normalization_profile') <> 'string'
           OR pg_catalog.jsonb_typeof(field.value -> 'state') <> 'string'
           OR pg_catalog.jsonb_typeof(field.value -> 'validation') <> 'object'
           OR EXISTS (
                SELECT 1
                FROM pg_catalog.jsonb_object_keys(field.value) AS key
                WHERE key NOT IN (
                    'id', 'key', 'logical_type', 'cardinality', 'required',
                    'classification', 'normalization_profile', 'state', 'validation'
                )
           )
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile schema field definition is invalid';
    END IF;
    SELECT active.id, active.version
    INTO v_active_id, v_active_version
    FROM iam_profile_schema_versions AS active
    WHERE active.resource_kind = p_resource_kind AND active.state = 'active'
    FOR UPDATE;
    IF NOT FOUND OR v_active_id IS DISTINCT FROM p_expected_active_schema_id OR
       p_version IS DISTINCT FROM v_active_version + 1 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'IAM profile active schema compare-and-swap failed';
    END IF;
    IF p_retire_field_id IS NOT NULL THEN
        IF p_expected_field_revision < 1 OR NOT EXISTS (
            SELECT 1
            FROM pg_catalog.jsonb_array_elements(p_definition -> 'fields') AS field(value)
            WHERE field.value ->> 'id' = p_retire_field_id::text
              AND field.value ->> 'state' = 'retired'
        ) THEN
            RAISE EXCEPTION USING ERRCODE = '22023',
                MESSAGE = 'IAM profile retirement target is invalid';
        END IF;
        UPDATE iam_profile_fields
        SET state = 'retired', retired_at = v_now, updated_at = v_now,
            revision = revision + 1
        WHERE id = p_retire_field_id AND resource_kind = p_resource_kind
          AND state = 'active' AND revision = p_expected_field_revision;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'IAM profile field retirement compare-and-swap failed';
        END IF;
    ELSIF p_expected_field_revision IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'IAM profile publish retirement evidence is invalid';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(p_definition -> 'fields') AS field(value)
        LEFT JOIN iam_profile_fields AS stored
          ON stored.id = (field.value ->> 'id')::uuid
         AND stored.resource_kind = p_resource_kind
        WHERE stored.id IS NULL AND field.value ->> 'state' <> 'active'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile new field must begin as active';
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
        SELECT 1
        FROM pg_catalog.jsonb_array_elements(p_definition -> 'fields') AS field(value)
        LEFT JOIN iam_profile_fields AS stored
          ON stored.id = (field.value ->> 'id')::uuid
         AND stored.resource_kind = p_resource_kind
        WHERE stored.id IS NULL OR stored.field_key IS DISTINCT FROM field.value ->> 'key'
           OR stored.value_kind IS DISTINCT FROM field.value ->> 'logical_type'
           OR stored.cardinality IS DISTINCT FROM field.value ->> 'cardinality'
           OR stored.required IS DISTINCT FROM (field.value ->> 'required')::boolean
           OR stored.classification IS DISTINCT FROM field.value ->> 'classification'
           OR stored.normalization_profile IS DISTINCT FROM
                field.value ->> 'normalization_profile'
           OR stored.state IS DISTINCT FROM field.value ->> 'state'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile stable field semantic is inconsistent';
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
        field.value -> 'validation',
        pg_catalog.sha256(pg_catalog.convert_to(
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
    v_definition_sha256 := iam_profile_definition_sha256_v1(p_schema_id);
    IF v_definition_sha256 IS DISTINCT FROM p_definition_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            MESSAGE = 'IAM profile schema definition hash is inconsistent';
    END IF;
    UPDATE iam_profile_schema_versions
    SET definition_sha256 = v_definition_sha256, updated_at = v_now,
        revision = revision + 1
    WHERE id = p_schema_id AND state = 'draft';
    UPDATE iam_profile_schema_versions
    SET state = 'retired', retired_at = v_now, updated_at = v_now,
        revision = revision + 1
    WHERE id = v_active_id AND state = 'active';
    UPDATE iam_profile_schema_versions
    SET state = 'active', activated_at = v_now, updated_at = v_now,
        revision = revision + 1
    WHERE id = p_schema_id AND state = 'draft'
    RETURNING revision INTO v_revision;
    IF v_revision IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'IAM profile schema activation failed';
    END IF;
    RETURN v_revision;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION iam_profile_schema_apply_v1(
    name,uuid,text,bigint,jsonb,bytea,uuid,uuid,bigint
) FROM PUBLIC;

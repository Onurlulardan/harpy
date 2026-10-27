-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_publish_v3(
    p_scope_schema name,
    p_revision_id uuid,
    p_expected_row_revision bigint,
    p_source_definition_sha256 bytea,
    p_compiler_version text,
    p_catalog_contract text,
    p_catalog_json text,
    p_catalog_sha256 bytea,
    p_presentation_contract text,
    p_presentation_json text,
    p_presentation_sha256 bytea,
    p_now timestamptz
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_catalog jsonb;
    v_catalog_ids jsonb;
    v_definition text;
    v_definition_sha256 bytea;
    v_field_ids jsonb;
    v_lock_id text;
    v_presentation jsonb;
    v_resource_kind text;
    v_row_revision bigint;
    v_state text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_revision_id IS NULL OR p_expected_row_revision < 1 OR p_now IS NULL
       OR p_compiler_version IS DISTINCT FROM
            'system-resource-presentation-compiler.v1'
       OR p_catalog_contract IS DISTINCT FROM
            'forms-presentation-field-catalog.v1'
       OR p_presentation_contract IS DISTINCT FROM 'forms-presentation.v1'
       OR pg_catalog.octet_length(p_source_definition_sha256) <> 32
       OR pg_catalog.octet_length(p_catalog_sha256) <> 32
       OR pg_catalog.octet_length(p_presentation_sha256) <> 32
       OR p_catalog_sha256 IS DISTINCT FROM pg_catalog.sha256(
            pg_catalog.convert_to(p_catalog_json, 'UTF8')
       )
       OR p_presentation_sha256 IS DISTINCT FROM pg_catalog.sha256(
            pg_catalog.convert_to(p_presentation_json, 'UTF8')
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_v3_contract_guard',
            MESSAGE = 'system resource presentation publish contract is invalid';
    END IF;
    BEGIN
        v_catalog := p_catalog_json::jsonb;
        v_presentation := p_presentation_json::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_v3_document_guard',
            MESSAGE = 'system resource presentation document is invalid';
    END;
    IF pg_catalog.jsonb_typeof(v_catalog) <> 'object'
       OR v_catalog ->> 'contract' IS DISTINCT FROM p_catalog_contract
       OR pg_catalog.jsonb_typeof(v_catalog -> 'fields') <> 'array'
       OR pg_catalog.jsonb_array_length(v_catalog -> 'fields') > 500
       OR pg_catalog.jsonb_typeof(v_presentation) <> 'object'
       OR v_presentation ->> 'registry_contract' IS DISTINCT FROM
            'forms-presentation-registry.v1'
       OR EXISTS (
            SELECT 1
            FROM pg_catalog.jsonb_array_elements(v_catalog -> 'fields') AS field(value)
            WHERE pg_catalog.jsonb_typeof(field.value) <> 'object'
               OR field.value ->> 'id' !~
                  '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
               OR field.value ->> 'id' = '00000000-0000-0000-0000-000000000000'
               OR field.value ->> 'mode' NOT IN ('editable', 'read_only')
               OR COALESCE(field.value ->> 'type_id', '') = ''
               OR COALESCE(field.value ->> 'value_profile', '') = ''
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_v3_document_guard',
            MESSAGE = 'system resource presentation document is invalid';
    END IF;
    SELECT COALESCE(
        pg_catalog.jsonb_agg(field.value -> 'id' ORDER BY field.value ->> 'id'),
        '[]'::jsonb
    ) INTO v_catalog_ids
    FROM pg_catalog.jsonb_array_elements(v_catalog -> 'fields') AS field(value);
    IF pg_catalog.jsonb_array_length(v_catalog_ids) IS DISTINCT FROM (
        SELECT pg_catalog.count(DISTINCT field.value #>> '{}')
        FROM pg_catalog.jsonb_array_elements(v_catalog_ids) AS field(value)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_v3_document_guard',
            MESSAGE = 'system resource presentation catalog contains duplicates';
    END IF;
    SELECT revision.state, revision.row_revision, revision.resource_kind,
           revision.definition_json, revision.definition_sha256
    INTO v_state, v_row_revision, v_resource_kind,
         v_definition, v_definition_sha256
    FROM system_resource_layout_revisions AS revision
    WHERE revision.id = p_revision_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'system resource layout revision was not found';
    END IF;
    IF v_state IS DISTINCT FROM 'draft'
       OR v_row_revision IS DISTINCT FROM p_expected_row_revision
       OR v_definition_sha256 IS DISTINCT FROM p_source_definition_sha256 THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'system resource layout revision conflicts';
    END IF;
    SELECT COALESCE(
        pg_catalog.jsonb_agg(field.value -> 'field_id' ORDER BY field.value ->> 'field_id'),
        '[]'::jsonb
    ) INTO v_field_ids
    FROM pg_catalog.jsonb_array_elements(v_definition::jsonb -> 'sections') section(value)
    CROSS JOIN pg_catalog.jsonb_array_elements(section.value -> 'fields') field(value);
    IF EXISTS (
        SELECT 1 FROM pg_catalog.jsonb_array_elements(v_field_ids) field(value)
        WHERE NOT v_catalog_ids ? (field.value #>> '{}')
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_v3_field_guard',
            MESSAGE = 'system resource layout fields are outside the supplied catalog';
    END IF;
    FOR v_lock_id IN
        SELECT referenced.field_id FROM (
            SELECT field.value ->> 'field_id' AS field_id
            FROM pg_catalog.jsonb_array_elements(v_definition::jsonb -> 'sections') section(value)
            CROSS JOIN pg_catalog.jsonb_array_elements(section.value -> 'fields') field(value)
            UNION
            SELECT field.value ->> 'field_id'
            FROM system_resource_layout_publications publication
            JOIN system_resource_layout_revisions published
              ON published.id = publication.revision_id
            CROSS JOIN pg_catalog.jsonb_array_elements(published.definition_json::jsonb -> 'sections') section(value)
            CROSS JOIN pg_catalog.jsonb_array_elements(section.value -> 'fields') field(value)
            WHERE publication.resource_kind = v_resource_kind
        ) referenced ORDER BY referenced.field_id
    LOOP
        PERFORM pg_catalog.pg_advisory_xact_lock(
            pg_catalog.hashtextextended(v_lock_id, 8501)
        );
    END LOOP;
    UPDATE system_resource_layout_revisions
    SET state = 'published', published_at = p_now,
        updated_at = p_now, row_revision = row_revision + 1
    WHERE id = p_revision_id;
    INSERT INTO system_resource_layout_publications (
        resource_kind, revision_id, generation, created_at, updated_at,
        source_definition_sha256, compiler_version, catalog_contract,
        catalog_json, catalog_sha256, presentation_contract,
        presentation_json, presentation_sha256
    ) VALUES (
        v_resource_kind, p_revision_id, 1, p_now, p_now,
        p_source_definition_sha256, p_compiler_version, p_catalog_contract,
        p_catalog_json, p_catalog_sha256, p_presentation_contract,
        p_presentation_json, p_presentation_sha256
    ) ON CONFLICT (resource_kind) DO UPDATE SET
        revision_id = EXCLUDED.revision_id,
        generation = system_resource_layout_publications.generation + 1,
        updated_at = EXCLUDED.updated_at,
        source_definition_sha256 = EXCLUDED.source_definition_sha256,
        compiler_version = EXCLUDED.compiler_version,
        catalog_contract = EXCLUDED.catalog_contract,
        catalog_json = EXCLUDED.catalog_json,
        catalog_sha256 = EXCLUDED.catalog_sha256,
        presentation_contract = EXCLUDED.presentation_contract,
        presentation_json = EXCLUDED.presentation_json,
        presentation_sha256 = EXCLUDED.presentation_sha256;
    RETURN system_resource_layout_publication_json_v2(v_resource_kind);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_publish_v3(
    name,uuid,bigint,bytea,text,text,text,bytea,text,text,bytea,timestamptz
) FROM PUBLIC;

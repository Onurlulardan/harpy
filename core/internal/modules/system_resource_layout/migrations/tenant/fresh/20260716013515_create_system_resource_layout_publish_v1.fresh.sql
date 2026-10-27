-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_publish_v1(
    p_scope_schema name,
    p_revision_id uuid,
    p_expected_row_revision bigint,
    p_catalog_json text,
    p_now timestamptz
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_catalog_document jsonb;
    v_catalog_ids jsonb;
    v_definition text;
    v_field_ids jsonb;
    v_resource_kind text;
    v_row_revision bigint;
    v_state text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_expected_row_revision < 1
       OR p_now IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_contract_guard',
            MESSAGE = 'system resource layout publish contract is invalid';
    END IF;
    BEGIN
        v_catalog_document := p_catalog_json::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_catalog_guard',
            MESSAGE = 'system resource layout catalog is invalid';
    END;
    IF pg_catalog.jsonb_typeof(v_catalog_document) <> 'array'
       OR pg_catalog.jsonb_array_length(v_catalog_document) > 512
       OR EXISTS (
            SELECT 1
            FROM pg_catalog.jsonb_array_elements(v_catalog_document) AS field(value)
            WHERE pg_catalog.jsonb_typeof(field.value) <> 'string'
               OR field.value #>> '{}' !~
                  '^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
               OR field.value #>> '{}' =
                  '00000000-0000-0000-0000-000000000000'
       ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_catalog_guard',
            MESSAGE = 'system resource layout catalog is invalid';
    END IF;
    SELECT COALESCE(
        pg_catalog.jsonb_agg(field.value ORDER BY field.value #>> '{}'),
        '[]'::jsonb
    )
    INTO v_catalog_ids
    FROM pg_catalog.jsonb_array_elements(v_catalog_document) AS field(value);
    IF pg_catalog.jsonb_array_length(v_catalog_ids) IS DISTINCT FROM (
        SELECT pg_catalog.count(DISTINCT catalog.value #>> '{}')
        FROM pg_catalog.jsonb_array_elements(v_catalog_ids) AS catalog(value)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_catalog_guard',
            MESSAGE = 'system resource layout catalog contains duplicates';
    END IF;
    SELECT revision.state, revision.row_revision, revision.resource_kind,
           revision.definition_json
    INTO v_state, v_row_revision, v_resource_kind, v_definition
    FROM system_resource_layout_revisions AS revision
    WHERE revision.id = p_revision_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'system resource layout revision was not found';
    END IF;
    IF v_state IS DISTINCT FROM 'draft'
       OR v_row_revision IS DISTINCT FROM p_expected_row_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'system resource layout revision conflicts';
    END IF;
    SELECT COALESCE(
        pg_catalog.jsonb_agg(
            field.value -> 'field_id'
            ORDER BY field.value ->> 'field_id'
        ),
        '[]'::jsonb
    )
    INTO v_field_ids
    FROM pg_catalog.jsonb_array_elements(
        v_definition::jsonb -> 'sections'
    ) AS section(value)
    CROSS JOIN pg_catalog.jsonb_array_elements(
        section.value -> 'fields'
    ) AS field(value);
    IF v_field_ids IS DISTINCT FROM v_catalog_ids THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_field_guard',
            MESSAGE = 'system resource layout fields do not match IAM catalog';
    END IF;
    UPDATE system_resource_layout_revisions
    SET state = 'published', published_at = p_now,
        updated_at = p_now, row_revision = row_revision + 1
    WHERE id = p_revision_id;
    INSERT INTO system_resource_layout_publications (
        resource_kind, revision_id, generation, created_at, updated_at
    ) VALUES (
        v_resource_kind, p_revision_id, 1, p_now, p_now
    )
    ON CONFLICT (resource_kind) DO UPDATE
    SET revision_id = EXCLUDED.revision_id,
        generation = system_resource_layout_publications.generation + 1,
        updated_at = EXCLUDED.updated_at;
    RETURN system_resource_layout_revision_json_v1(p_revision_id);
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_publish_v1(
    name,uuid,bigint,text,timestamptz
) FROM PUBLIC;

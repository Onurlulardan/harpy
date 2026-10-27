-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_publish_v5(
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
    v_resource_kind text;
    v_row_revision bigint;
    v_state text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'forgecrud_owner'
       OR p_scope_schema IS DISTINCT FROM pg_catalog.current_schema()
       OR p_revision_id IS NULL OR p_expected_row_revision < 1
       OR p_expected_row_revision = 9223372036854775807 OR p_now IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '23514',
            CONSTRAINT = 'system_resource_layout_publish_v5_contract_guard',
            MESSAGE = 'system resource layout v5 publish contract is invalid';
    END IF;
    SELECT revision.state, revision.row_revision, revision.resource_kind
    INTO v_state, v_row_revision, v_resource_kind
    FROM system_resource_layout_revisions AS revision
    WHERE revision.id = p_revision_id
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'system resource layout revision was not found';
    END IF;
    IF v_state = 'published' THEN
        IF v_row_revision IS DISTINCT FROM p_expected_row_revision + 1
           OR NOT EXISTS (
                SELECT 1
                FROM system_resource_layout_publications AS publication
                WHERE publication.resource_kind = v_resource_kind
                  AND publication.revision_id = p_revision_id
                  AND publication.source_definition_sha256 = p_source_definition_sha256
                  AND publication.compiler_version = p_compiler_version
                  AND publication.catalog_contract = p_catalog_contract
                  AND publication.catalog_json = p_catalog_json
                  AND publication.catalog_sha256 = p_catalog_sha256
                  AND publication.presentation_contract = p_presentation_contract
                  AND publication.presentation_json = p_presentation_json
                  AND publication.presentation_sha256 = p_presentation_sha256
           ) THEN
            RAISE EXCEPTION USING ERRCODE = '40001',
                MESSAGE = 'system resource layout publication replay conflicts';
        END IF;
        RETURN system_resource_layout_publication_json_v2(v_resource_kind);
    END IF;
    IF v_state IS DISTINCT FROM 'draft'
       OR v_row_revision IS DISTINCT FROM p_expected_row_revision THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'system resource layout revision conflicts';
    END IF;
    RETURN system_resource_layout_publish_v4(
        p_scope_schema, p_revision_id, p_expected_row_revision,
        p_source_definition_sha256, p_compiler_version,
        p_catalog_contract, p_catalog_json, p_catalog_sha256,
        p_presentation_contract, p_presentation_json,
        p_presentation_sha256, p_now
    );
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_publish_v5(
    name,uuid,bigint,bytea,text,text,text,bytea,text,text,bytea,timestamptz
) FROM PUBLIC;

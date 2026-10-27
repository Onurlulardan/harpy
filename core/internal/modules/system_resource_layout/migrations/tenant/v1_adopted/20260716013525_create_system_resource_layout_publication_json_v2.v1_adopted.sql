-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_publication_json_v2(
    p_resource_kind text
)
RETURNS jsonb
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_resource_kind NOT IN ('user', 'organization', 'role') THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'system resource layout publication read is invalid';
    END IF;
    SELECT pg_catalog.jsonb_build_object(
        'revision', system_resource_layout_revision_json_v1(publication.revision_id),
        'generation', publication.generation,
        'source_definition_sha256', pg_catalog.encode(
            publication.source_definition_sha256, 'hex'
        ),
        'compiler_version', publication.compiler_version,
        'catalog_contract', publication.catalog_contract,
        'catalog_json', publication.catalog_json,
        'catalog_sha256', pg_catalog.encode(publication.catalog_sha256, 'hex'),
        'presentation_contract', publication.presentation_contract,
        'presentation_json', publication.presentation_json,
        'presentation_sha256', pg_catalog.encode(
            publication.presentation_sha256, 'hex'
        )
    )
    INTO v_document
    FROM system_resource_layout_publications AS publication
    WHERE publication.resource_kind = p_resource_kind;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P0002',
            MESSAGE = 'published system resource layout was not found';
    END IF;
    RETURN v_document;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION system_resource_layout_publication_json_v2(text)
    FROM PUBLIC;

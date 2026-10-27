-- +goose Up
-- Migration stream version: 20260802232101.
-- +goose StatementBegin
CREATE FUNCTION system_resource_layout_curated_successor_catalog_v1(
    p_resource_kind text, p_catalog_json text
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_catalog jsonb; v_fields jsonb;
BEGIN
    IF NOT system_resource_layout_installation_catalog_valid_v2(
        p_resource_kind,p_catalog_json) THEN
        RAISE EXCEPTION USING ERRCODE='23514',
            MESSAGE='system resource curated predecessor catalog is invalid';
    END IF;
    IF p_resource_kind <> 'organization' THEN RETURN p_catalog_json; END IF;
    v_catalog:=p_catalog_json::jsonb;
    WITH expanded AS (
        SELECT field.value FROM pg_catalog.jsonb_array_elements(
            v_catalog->'fields') AS field(value)
        UNION ALL
        SELECT pg_catalog.jsonb_build_object(
            'id','c1000000-0000-4000-8000-00000000000c',
            'mode','editable','type_id','uuid_list',
            'value_profile','uuid_list_v1')
    )
    SELECT pg_catalog.jsonb_agg(value ORDER BY value->>'id')
      INTO v_fields FROM expanded;
    v_catalog:=pg_catalog.jsonb_set(v_catalog,'{fields}',v_fields,false);
    RETURN system_resource_layout_canonical_json_v1(v_catalog);
END;
$function$;

CREATE FUNCTION system_resource_layout_curated_predecessor_valid_v1(
    p_resource_kind text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_publication system_resource_layout_publications%ROWTYPE;
    v_revision system_resource_layout_revisions%ROWTYPE;
    v_definition text; v_presentation text;
BEGIN
    IF p_resource_kind NOT IN ('user','organization','role')
       OR (SELECT pg_catalog.count(*) FROM
           system_resource_layout_installation_baselines)<>1 THEN
        RETURN false;
    END IF;
    SELECT publication.* INTO v_publication
      FROM system_resource_layout_publications AS publication
     WHERE publication.resource_kind=p_resource_kind;
    IF NOT FOUND OR v_publication.generation<>2
       OR v_publication.catalog_contract<>
          'forms-presentation-field-catalog.v1'
       OR v_publication.presentation_contract<>'forms-presentation.v1'
       OR v_publication.compiler_version<>
          'system-resource-presentation-compiler.v2'
       OR v_publication.catalog_sha256 IS DISTINCT FROM pg_catalog.sha256(
          pg_catalog.convert_to(v_publication.catalog_json,'UTF8'))
       OR v_publication.presentation_sha256 IS DISTINCT FROM pg_catalog.sha256(
          pg_catalog.convert_to(v_publication.presentation_json,'UTF8'))
       OR NOT system_resource_layout_installation_catalog_valid_v2(
          p_resource_kind,v_publication.catalog_json) THEN RETURN false; END IF;
    SELECT revision.* INTO v_revision
      FROM system_resource_layout_revisions AS revision
     WHERE revision.id=v_publication.revision_id;
    IF NOT FOUND OR v_revision.resource_kind<>p_resource_kind
       OR v_revision.state<>'published' OR v_revision.row_revision<>2
       OR v_revision.definition_sha256 IS DISTINCT FROM pg_catalog.sha256(
          pg_catalog.convert_to(v_revision.definition_json,'UTF8'))
       OR v_publication.source_definition_sha256 IS DISTINCT FROM
          v_revision.definition_sha256 THEN RETURN false; END IF;
    v_definition:=system_resource_layout_curated_definition_v1(
        p_resource_kind,v_publication.catalog_json);
    IF v_revision.definition_json IS DISTINCT FROM v_definition THEN
        RETURN false;
    END IF;
    v_presentation:=system_resource_layout_compile_curated_v1(
        v_revision.id,v_revision.created_at,v_definition,
        v_publication.catalog_json);
    RETURN v_publication.presentation_json IS NOT DISTINCT FROM v_presentation;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    system_resource_layout_curated_successor_catalog_v1(text,text)
    FROM PUBLIC;
REVOKE ALL ON FUNCTION
    system_resource_layout_curated_predecessor_valid_v1(text)
    FROM PUBLIC;

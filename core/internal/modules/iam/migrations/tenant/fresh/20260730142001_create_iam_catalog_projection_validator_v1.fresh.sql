-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION iam_catalog_project_document_v1(
    p_package_slug text, p_component_id uuid,
    p_contract text, p_content_sha256 text,
    p_component_kind text, p_module_slug text,
    p_semantic_identity text, p_semantic_sha256 bytea,
    p_item_count integer, p_auxiliary_count integer,
    p_document_bytes bytea
) RETURNS jsonb
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_document jsonb;
BEGIN
    IF p_package_slug !~ '^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$' OR
       p_module_slug IS DISTINCT FROM p_package_slug OR
       p_semantic_identity IS DISTINCT FROM p_package_slug OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_contract IS DISTINCT FROM 'iam.catalog-template.v1' OR
       p_content_sha256 !~ '^[0-9a-f]{64}$' OR
       p_component_kind IS DISTINCT FROM 'IAM_TEMPLATE' OR
       octet_length(p_semantic_sha256) IS DISTINCT FROM 32 OR
       p_item_count < 2 OR p_auxiliary_count < 1 OR
       p_document_bytes IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM catalog document input is invalid';
    END IF;
    IF pg_catalog.encode(pg_catalog.sha256(p_document_bytes), 'hex')
           IS DISTINCT FROM p_content_sha256 OR
       p_semantic_sha256 IS DISTINCT FROM pg_catalog.sha256(p_document_bytes) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM catalog document digest differs';
    END IF;
    BEGIN
        v_document := pg_catalog.convert_from(p_document_bytes, 'UTF8')::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM catalog document is invalid';
    END;
    IF jsonb_typeof(v_document) IS DISTINCT FROM 'object' OR
       v_document->>'contract_version' IS DISTINCT FROM p_contract OR
       v_document->>'source_contract'
           IS DISTINCT FROM 'harpy-v1.enterprise-seed.' || p_package_slug OR
       v_document->>'component_id' IS DISTINCT FROM p_component_id::text OR
       v_document->>'module_slug' IS DISTINCT FROM p_package_slug OR
       jsonb_typeof(v_document->'resources') IS DISTINCT FROM 'array' OR
       jsonb_typeof(v_document->'grants') IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM catalog document binding differs';
    END IF;
    IF jsonb_array_length(v_document->'resources') <> p_item_count OR
       jsonb_array_length(v_document->'grants') <> p_auxiliary_count OR
       (SELECT count(*) FROM jsonb_to_recordset(v_document->'resources')
            AS resource(slug text, kind text, name text, description text)
        WHERE resource.slug = 'module.' || p_package_slug
          AND resource.kind = 'module') <> 1 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM catalog document composition differs';
    END IF;
    RETURN v_document;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION iam_catalog_project_document_v1(
    text,uuid,text,text,text,text,text,bytea,integer,integer,bytea
) FROM PUBLIC;

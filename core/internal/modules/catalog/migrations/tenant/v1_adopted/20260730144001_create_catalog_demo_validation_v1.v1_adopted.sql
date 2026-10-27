-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION catalog_demo_assert_current_authority_v1(
    p_scope_schema name, p_generation bigint
) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_schema text;
    v_tenant_id uuid;
BEGIN
    v_schema := pg_catalog.current_schema()::text;
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       p_scope_schema IS NULL OR p_scope_schema::text IS DISTINCT FROM v_schema OR
       p_generation < 1 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog demo runtime authority role is invalid';
    END IF;
    BEGIN
        SELECT identity_row.tenant_id INTO STRICT v_tenant_id
        FROM tenancy_schema_identity AS identity_row
        WHERE identity_row.singleton AND identity_row.schema_name = v_schema;
    EXCEPTION WHEN NO_DATA_FOUND OR TOO_MANY_ROWS THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Catalog demo TenantScope identity is invalid';
    END;
    PERFORM platform.catalog_assert_tenant_authority_v1(
        v_tenant_id, p_generation
    );
    RETURN v_tenant_id;
END;
$function$;

CREATE FUNCTION catalog_demo_project_document_v1(
    p_package_slug text, p_component_id uuid, p_contract text,
    p_content_sha256 text, p_module_slug text, p_source_contract text,
    p_semantic_sha256 bytea, p_form_count integer, p_row_count integer,
    p_document_bytes bytea
) RETURNS jsonb
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_document jsonb;
    v_rows bigint;
    v_forms_valid boolean;
BEGIN
    IF p_package_slug IS NULL OR
       p_package_slug !~ '^[a-z][a-z0-9]*(?:-[a-z0-9]+)*$' OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_contract IS DISTINCT FROM 'catalog.demo-data.v1' OR
       p_content_sha256 !~ '^[0-9a-f]{64}$' OR
       p_module_slug IS DISTINCT FROM p_package_slug OR
       p_source_contract IS DISTINCT FROM
           'harpy-v1.enterprise-seed.' || p_package_slug OR
       octet_length(p_semantic_sha256) IS DISTINCT FROM 32 OR
       p_semantic_sha256 IS DISTINCT FROM pg_catalog.sha256(p_document_bytes) OR
       pg_catalog.encode(pg_catalog.sha256(p_document_bytes), 'hex')
           IS DISTINCT FROM p_content_sha256 OR
       p_form_count < 1 OR p_row_count < 1 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Catalog demo payload identity is invalid';
    END IF;
    BEGIN
        v_document := pg_catalog.convert_from(p_document_bytes, 'UTF8')::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Catalog demo payload JSON is invalid';
    END;
    IF jsonb_typeof(v_document) IS DISTINCT FROM 'object' OR
       v_document->>'contract_version' IS DISTINCT FROM p_contract OR
       v_document->>'component_id' IS DISTINCT FROM p_component_id::text OR
       v_document->>'module_slug' IS DISTINCT FROM p_package_slug OR
       v_document->>'source_contract' IS DISTINCT FROM p_source_contract OR
       jsonb_typeof(v_document->'demo_forms') IS DISTINCT FROM 'array' OR
       jsonb_array_length(v_document->'demo_forms') IS DISTINCT FROM p_form_count THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Catalog demo payload contract is invalid';
    END IF;
    SELECT bool_and(
        jsonb_typeof(form.value) = 'object' AND
        jsonb_typeof(form.value->'rows') = 'array'
    ), sum(CASE WHEN jsonb_typeof(form.value->'rows') = 'array'
        THEN jsonb_array_length(form.value->'rows') ELSE 0 END)
    INTO v_forms_valid, v_rows
    FROM jsonb_array_elements(v_document->'demo_forms') AS form(value);
    IF v_forms_valid IS DISTINCT FROM true OR v_rows IS DISTINCT FROM p_row_count THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Catalog demo payload counts are invalid';
    END IF;
    RETURN v_document;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION catalog_demo_assert_current_authority_v1(name,bigint)
    FROM PUBLIC;
REVOKE ALL ON FUNCTION catalog_demo_project_document_v1(
    text,uuid,text,text,text,text,bytea,integer,integer,bytea
) FROM PUBLIC;

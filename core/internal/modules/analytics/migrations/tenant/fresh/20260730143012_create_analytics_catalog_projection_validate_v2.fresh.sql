-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_projection_validate_v2(
    p_contract text,p_content_sha256 text,p_component_id uuid,
    p_component_kind text,p_module_slug text,p_semantic_identity text,
    p_semantic_sha256 bytea,p_item_count integer,p_auxiliary_count integer,
    p_document_bytes bytea,p_runtime_plan_bytes bytea,
    p_runtime_plan_sha256 bytea
)
RETURNS TABLE(
    analytics_kind text,name text,description text,legacy_hash text,
    definition_json jsonb,definition_sha256 bytea,blocker_code text
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_original jsonb;
    v_plan jsonb;
    v_kind text;
    v_auxiliary integer;
    v_blocker text;
    v_expected_sources jsonb;
    v_valid_sources bigint;
BEGIN
    IF p_contract IS DISTINCT FROM 'analytics.catalog-definition.v1' OR
       p_content_sha256 IS NULL OR
       p_content_sha256 !~ '^[0-9a-f]{64}$' OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_component_kind IS NULL OR
       p_component_kind NOT IN ('ENDPOINT','DASHBOARD') OR
       p_module_slug IS NULL OR
       p_module_slug !~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$' OR
       p_semantic_identity IS NULL OR
       p_semantic_identity !~ '^[a-z][a-z0-9-]{1,62}$' OR
       pg_catalog.octet_length(p_semantic_sha256) IS DISTINCT FROM 32 OR
       p_item_count IS DISTINCT FROM 1 OR p_auxiliary_count<2 OR
       pg_catalog.octet_length(p_runtime_plan_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.encode(pg_catalog.sha256(p_document_bytes),'hex')
           IS DISTINCT FROM p_content_sha256 OR
       pg_catalog.sha256(p_runtime_plan_bytes)
           IS DISTINCT FROM p_runtime_plan_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection document input is invalid';
    END IF;
    BEGIN
        v_original:=pg_catalog.convert_from(p_document_bytes,'UTF8')::jsonb;
        v_plan:=pg_catalog.convert_from(p_runtime_plan_bytes,'UTF8')::jsonb;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection document is invalid';
    END;
    v_kind:=v_original->>'kind';
    IF pg_catalog.jsonb_typeof(v_original) IS DISTINCT FROM 'object' OR
       v_original->>'contract_version' IS DISTINCT FROM p_contract OR
       v_original->>'component_id' IS DISTINCT FROM p_component_id::text OR
       v_original->>'analytics_id' IS DISTINCT FROM p_component_id::text OR
       v_original->>'source_contract' IS DISTINCT FROM
           'harpy-v1.enterprise-seed.'||p_module_slug OR
       v_original->>'slug' IS DISTINCT FROM p_semantic_identity OR
       v_original->>'name' IS NULL OR
       pg_catalog.octet_length(v_original->>'name') NOT BETWEEN 1 AND 100 OR
       v_original->>'description' IS NULL OR
       pg_catalog.octet_length(v_original->>'description')
           NOT BETWEEN 1 AND 2048 OR
       v_original->>'legacy_hash' IS NULL OR
       v_original->>'legacy_hash' !~ '^[0-9a-f]{64}$' OR
       v_original->>'legacy_hash'
           IS DISTINCT FROM pg_catalog.encode(p_semantic_sha256,'hex') OR
       v_kind IS NULL OR v_kind NOT IN ('endpoint','dashboard') OR
       (v_kind='endpoint' AND p_component_kind<>'ENDPOINT') OR
       (v_kind='dashboard' AND p_component_kind<>'DASHBOARD') OR
       pg_catalog.jsonb_typeof(v_original#>'{definition,dataSources}')
           IS DISTINCT FROM 'array' OR
       v_plan->>'contract'
           IS DISTINCT FROM 'analytics.catalog-runtime-plan.v1' OR
       v_plan->>'kind' IS DISTINCT FROM v_kind OR
       v_plan->>'component_id' IS DISTINCT FROM p_component_id::text OR
       v_plan->>'source_contract' IS DISTINCT FROM
           v_original->>'source_contract' OR
       v_plan->>'document_sha256' IS DISTINCT FROM p_content_sha256 OR
       v_plan->>'status' IS NULL OR
       v_plan->>'status' NOT IN ('candidate','pending') OR
       pg_catalog.jsonb_typeof(v_plan->'sources') IS DISTINCT FROM 'array' THEN
       RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection binding differs';
    END IF;
    IF v_kind='endpoint' THEN
        IF pg_catalog.jsonb_typeof(v_original#>'{definition,fields}')
               IS DISTINCT FROM 'array' OR
           pg_catalog.jsonb_array_length(
               v_original#>'{definition,dataSources}'
           )<1 OR
           pg_catalog.jsonb_array_length(
               v_original#>'{definition,fields}'
           )<1 THEN
            RAISE EXCEPTION USING ERRCODE='22023',
                MESSAGE='Analytics Catalog endpoint is invalid';
        END IF;
        SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(
                   source.value->>'formSlug'
               ) ORDER BY source.ordinality),
               pg_catalog.count(*) FILTER (WHERE
                   pg_catalog.jsonb_typeof(source.value)='object' AND
                   source.value->>'formSlug' IS NOT NULL AND
                   source.value->>'formSlug'<>''
               )
        INTO v_expected_sources,v_valid_sources
        FROM pg_catalog.jsonb_array_elements(
            v_original#>'{definition,dataSources}'
        ) WITH ORDINALITY AS source(value,ordinality);
        IF v_expected_sources IS DISTINCT FROM v_plan->'sources' OR
           v_valid_sources<>pg_catalog.jsonb_array_length(
               v_original#>'{definition,dataSources}'
           ) OR pg_catalog.jsonb_typeof(v_plan->'parameters')
               IS DISTINCT FROM 'object' OR
           pg_catalog.jsonb_typeof(v_plan->'output')
               IS DISTINCT FROM 'object' OR
           (v_plan->>'status'='candidate' AND
            pg_catalog.jsonb_typeof(v_plan->'query')
                IS DISTINCT FROM 'object') OR
           (v_plan->>'status'='pending' AND v_plan ? 'query') OR
           v_plan ? 'dashboard' THEN
            RAISE EXCEPTION USING ERRCODE='22023',
                MESSAGE='Analytics Catalog endpoint plan differs';
        END IF;
        v_auxiliary:=
            pg_catalog.jsonb_array_length(v_original#>'{definition,dataSources}')+
            pg_catalog.jsonb_array_length(v_original#>'{definition,fields}');
    ELSE
        IF pg_catalog.jsonb_typeof(v_original#>'{definition,widgets}')
               IS DISTINCT FROM 'array' OR
           pg_catalog.jsonb_array_length(
               v_original#>'{definition,dataSources}'
           )<1 OR
           pg_catalog.jsonb_array_length(
               v_original#>'{definition,widgets}'
           )<1 THEN
            RAISE EXCEPTION USING ERRCODE='22023',
                MESSAGE='Analytics Catalog dashboard is invalid';
        END IF;
        SELECT pg_catalog.jsonb_agg(pg_catalog.to_jsonb(
                   source.value->>'endpointSlug'
               ) ORDER BY source.ordinality),
               pg_catalog.count(*) FILTER (WHERE
                   pg_catalog.jsonb_typeof(source.value)='object' AND
                   source.value->>'endpointSlug' IS NOT NULL AND
                   source.value->>'endpointSlug'<>''
               )
        INTO v_expected_sources,v_valid_sources
        FROM pg_catalog.jsonb_array_elements(
            v_original#>'{definition,dataSources}'
        ) WITH ORDINALITY AS source(value,ordinality);
        IF v_plan->>'status'<>'candidate' OR
           v_expected_sources IS DISTINCT FROM v_plan->'sources' OR
           v_valid_sources<>pg_catalog.jsonb_array_length(
               v_original#>'{definition,dataSources}'
           ) OR v_plan->'dashboard' IS DISTINCT FROM
               v_original->'definition' OR
           v_plan ? 'query' OR v_plan ? 'parameters' OR v_plan ? 'output' THEN
            RAISE EXCEPTION USING ERRCODE='22023',
                MESSAGE='Analytics Catalog dashboard plan differs';
        END IF;
        v_auxiliary:=
            pg_catalog.jsonb_array_length(v_original#>'{definition,dataSources}')+
            pg_catalog.jsonb_array_length(v_original#>'{definition,widgets}');
    END IF;
    IF v_auxiliary<>p_auxiliary_count OR
       (v_plan->>'status'='pending' AND (
        v_plan->>'blocker_code' IS NULL OR
        v_plan->>'blocker_code' !~ '^[A-Z][A-Z0-9_]{0,127}$'
       )) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog projection composition differs';
    END IF;
    v_blocker:=CASE WHEN v_plan->>'status'='pending'
        THEN v_plan->>'blocker_code'
        ELSE 'ANALYTICS_DEPENDENCY_NOT_READY' END;
    RETURN QUERY SELECT v_kind,v_original->>'name',
        v_original->>'description',v_original->>'legacy_hash',
        v_original->'definition',
        pg_catalog.sha256(pg_catalog.convert_to(
            (v_original->'definition')::text,'UTF8'
        )),v_blocker;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_projection_validate_v2(
    text,text,uuid,text,text,text,bytea,integer,integer,bytea,bytea,bytea
) FROM PUBLIC;

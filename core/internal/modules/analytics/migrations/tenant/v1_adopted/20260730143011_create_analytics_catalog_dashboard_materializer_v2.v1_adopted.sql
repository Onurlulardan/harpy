-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_materialize_dashboard_v2(
    p_scope_schema name, p_component_id uuid, p_version_seed uuid,
    p_package_id uuid, p_package_version text, p_release_id uuid,
    p_actor_id uuid, p_original jsonb, p_plan jsonb, p_occurred_at timestamptz
)
RETURNS TABLE(
    operational_state text, blocker_code text,
    operational_resource_id uuid, published_version_id uuid,
    current_draft_version_id uuid
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_dashboard_id uuid;
    v_published_id uuid;
    v_draft_id uuid;
    v_create_operation uuid;
    v_publish_operation uuid;
    v_layout_document bytea;
    v_widgets jsonb;
    v_expected_count bigint;
    v_resolved_count bigint;
    v_manifest_material bytea;
    v_manifest bytea;
    v_item jsonb;
    v_query_version_id uuid;
    v_query_manifest bytea;
    v_read_version uuid;
    v_read_draft uuid;
    v_read_manifest bytea;
    v_read_widgets jsonb;
    v_read_state text;
BEGIN
    v_manifest_material := ''::bytea;
    IF pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_version_seed) IS DISTINCT FROM 7 OR
       p_package_id IS NULL OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       p_plan->>'contract' <> 'analytics.catalog-runtime-plan.v1' OR
       p_plan->>'kind' <> 'dashboard' OR
       p_plan->>'status' NOT IN ('candidate','pending') THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog dashboard v2 plan is invalid';
    END IF;
    IF p_plan->>'status' = 'pending' THEN
        RETURN QUERY SELECT 'pending',p_plan->>'blocker_code',
            NULL::uuid,NULL::uuid,NULL::uuid;
        RETURN;
    END IF;
    IF pg_catalog.jsonb_typeof(p_plan->'dashboard') <> 'object' OR
       pg_catalog.jsonb_typeof(p_plan#>'{dashboard,dataSources}') <> 'array' OR
       pg_catalog.jsonb_typeof(p_plan#>'{dashboard,widgets}') <> 'array' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog dashboard v2 candidate is invalid';
    END IF;
    v_expected_count := pg_catalog.jsonb_array_length(
        p_plan#>'{dashboard,widgets}'
    );
    IF v_expected_count NOT BETWEEN 1 AND 30 THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog dashboard v2 widget count is invalid';
    END IF;
    SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'widget_id',analytics_catalog_uuid_v7_v1(
            p_version_seed,'widget:'||
            pg_catalog.lpad((widget.ordinality-1)::text,2,'0')
        ),
        'position',widget.ordinality-1,
        'query_version_id',endpoint.published_version_id,
        'parameter_document','7b7d','output_document','7b7d'
    ) ORDER BY widget.ordinality),pg_catalog.count(*)
    INTO v_widgets,v_resolved_count
    FROM pg_catalog.jsonb_array_elements(p_plan#>'{dashboard,widgets}')
         WITH ORDINALITY AS widget(value,ordinality)
    JOIN pg_catalog.jsonb_array_elements(p_plan#>'{dashboard,dataSources}')
         AS source(value)
      ON source.value->>'id'=widget.value->>'dataSource'
    JOIN analytics_catalog_version_projections AS endpoint
      ON endpoint.package_id=p_package_id
     AND endpoint.package_version=p_package_version
     AND endpoint.release_id=p_release_id
     AND endpoint.semantic_identity=source.value->>'endpointSlug'
     AND endpoint.analytics_kind='endpoint'
     AND endpoint.operational_state='ready'
     AND endpoint.ready_at IS NOT NULL;
    IF v_resolved_count <> v_expected_count THEN
        RETURN QUERY SELECT 'pending',
            'ANALYTICS_ENDPOINT_DEPENDENCY_NOT_READY',
            NULL::uuid,NULL::uuid,NULL::uuid;
        RETURN;
    END IF;
    v_dashboard_id := analytics_catalog_uuid_v7_v1(
        p_version_seed,'dashboard:root'
    );
    v_published_id := analytics_catalog_uuid_v7_v1(
        p_version_seed,'dashboard:published'
    );
    v_draft_id := analytics_catalog_uuid_v7_v1(
        p_version_seed,'dashboard:draft'
    );
    v_create_operation := analytics_catalog_uuid_v7_v1(
        p_version_seed,'dashboard:create-op'
    );
    v_publish_operation := analytics_catalog_uuid_v7_v1(
        p_version_seed,'dashboard:publish-op'
    );
    v_layout_document := pg_catalog.convert_to(
        (p_plan->'dashboard')::text,'UTF8'
    );
    FOR v_item IN
        SELECT item.value
        FROM pg_catalog.jsonb_array_elements(v_widgets) AS item(value)
        ORDER BY (item.value->>'position')::integer
    LOOP
        v_query_version_id := (v_item->>'query_version_id')::uuid;
        v_query_manifest := analytics_catalog_resource_manifest_v2(
            p_scope_schema,'query',v_query_version_id
        );
        v_manifest_material := v_manifest_material ||
            pg_catalog.uuid_send(v_query_version_id) || v_query_manifest;
    END LOOP;
    v_manifest := pg_catalog.sha256(v_manifest_material);
    PERFORM 1 FROM analytics_dashboard_create_v1(
        p_scope_schema,v_create_operation,p_actor_id,
        'catalog.analytics.'||p_version_seed::text||'.dashboard.create',
        pg_catalog.sha256(pg_catalog.convert_to(p_plan::text||':create','UTF8')),
        v_dashboard_id,v_published_id,p_original->>'name','tenant',
        v_layout_document,v_widgets,p_occurred_at
    );
    PERFORM 1 FROM analytics_dashboard_publish_v1(
        p_scope_schema,v_publish_operation,p_actor_id,
        'catalog.analytics.'||p_version_seed::text||'.dashboard.publish',
        pg_catalog.sha256(pg_catalog.convert_to(p_plan::text||':publish','UTF8')),
        v_dashboard_id,v_published_id,v_draft_id,1,1,v_manifest,p_occurred_at
    );
    SELECT result.version_id,result.version_state,result.manifest_sha256,
        result.widgets
    INTO v_read_version,v_read_state,v_read_manifest,v_read_widgets
    FROM analytics_dashboard_get_version_v1(
        p_scope_schema,v_dashboard_id,v_published_id
    ) AS result;
    SELECT current_dashboard.version_id INTO v_read_draft
    FROM analytics_dashboard_get_v1(
        p_scope_schema,v_dashboard_id
    ) AS current_dashboard;
    IF v_read_version <> v_published_id OR v_read_state <> 'published' OR
       v_read_draft <> v_draft_id OR v_read_manifest <> v_manifest OR
       v_read_widgets <> v_widgets OR
       analytics_catalog_resource_manifest_v2(
           p_scope_schema,'dashboard',v_published_id
       ) <> v_manifest THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog dashboard v2 readback differs';
    END IF;
    RETURN QUERY SELECT 'ready',NULL::text,
        v_dashboard_id,v_published_id,v_draft_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_materialize_dashboard_v2(
    name,uuid,uuid,uuid,text,uuid,uuid,jsonb,jsonb,timestamptz
) FROM PUBLIC;

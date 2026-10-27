-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_materialize_endpoint_v1(
    p_scope_schema name, p_component_id uuid, p_actor_id uuid,
    p_original jsonb, p_plan jsonb, p_occurred_at timestamptz
)
RETURNS TABLE(
    operational_state text, blocker_code text,
    operational_resource_id uuid, published_version_id uuid,
    current_draft_version_id uuid
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_source_count integer;
    v_resolved_count bigint;
    v_dependency_valid boolean;
    v_query_id uuid;
    v_published_id uuid;
    v_draft_id uuid;
    v_create_operation uuid;
    v_publish_operation uuid;
    v_query_document bytea;
    v_parameter_document bytea;
    v_output_document bytea;
    v_dependencies jsonb;
    v_manifest bytea;
    v_read_version uuid;
    v_read_draft uuid;
    v_read_manifest bytea;
    v_read_dependencies jsonb;
BEGIN
    IF p_plan->>'contract' <> 'analytics.catalog-runtime-plan.v1' OR
       p_plan->>'kind' <> 'endpoint' OR
       p_plan->>'status' NOT IN ('candidate','pending') THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog endpoint runtime plan is invalid';
    END IF;
    IF p_plan->>'status' = 'pending' THEN
        RETURN QUERY SELECT 'pending', p_plan->>'blocker_code',
            NULL::uuid, NULL::uuid, NULL::uuid;
        RETURN;
    END IF;
    IF pg_catalog.jsonb_typeof(p_plan->'sources') IS DISTINCT FROM 'array' OR
       pg_catalog.jsonb_array_length(p_plan->'sources') NOT BETWEEN 1 AND 4 OR
       pg_catalog.jsonb_typeof(p_plan->'query') IS DISTINCT FROM 'object' OR
       pg_catalog.jsonb_typeof(p_plan->'parameters') IS DISTINCT FROM 'object' OR
       pg_catalog.jsonb_typeof(p_plan->'output') IS DISTINCT FROM 'object' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog endpoint candidate is invalid';
    END IF;
    v_source_count := pg_catalog.jsonb_array_length(p_plan->'sources');
    IF (SELECT pg_catalog.count(DISTINCT source.value)
        FROM pg_catalog.jsonb_array_elements_text(p_plan->'sources')
            AS source(value)) <> v_source_count THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog endpoint sources differ';
    END IF;
    SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'owner_domain','forms','resource_kind','form',
        'resource_id',dependency.definition_id,
        'resource_version_id',dependency.version_id,
        'owner_generation',dependency.owner_generation,
        'catalog_version',dependency.catalog_version,
        'query_contract_version',dependency.query_contract_version,
        'policy_version',dependency.policy_version,
        'data_revision_capability',dependency.data_revision_capability,
        'dependency_sha256',pg_catalog.encode(
            dependency.dependency_sha256,'hex'
        )
    ) ORDER BY source.ordinality),pg_catalog.count(*),pg_catalog.bool_and(
        pg_catalog.uuid_extract_version(dependency.publication_id)=7 AND
        pg_catalog.uuid_extract_version(dependency.definition_id)=7 AND
        pg_catalog.uuid_extract_version(dependency.version_id)=7 AND
        dependency.owner_generation>0 AND dependency.catalog_version>0 AND
        dependency.query_contract_version>0 AND dependency.policy_version>0 AND
        pg_catalog.octet_length(dependency.data_revision_capability)
            BETWEEN 0 AND 128 AND
        pg_catalog.octet_length(dependency.dependency_sha256)=32
    ) INTO v_dependencies,v_resolved_count,v_dependency_valid
    FROM pg_catalog.jsonb_array_elements_text(p_plan->'sources')
        WITH ORDINALITY AS source(value,ordinality)
    JOIN forms_catalog_analytics_dependency_v1(
        p_scope_schema,source.value
    ) AS dependency ON TRUE;
    IF v_resolved_count <> v_source_count THEN
        RETURN QUERY SELECT 'pending',
            'ANALYTICS_FORMS_DEPENDENCY_NOT_READY',
            NULL::uuid,NULL::uuid,NULL::uuid;
        RETURN;
    END IF;
    IF v_dependency_valid IS NOT TRUE THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog endpoint dependency differs';
    END IF;
    v_query_id := analytics_catalog_uuid_v7_v1(p_component_id,'query:root');
    v_published_id := analytics_catalog_uuid_v7_v1(
        p_component_id,'query:published'
    );
    v_draft_id := analytics_catalog_uuid_v7_v1(p_component_id,'query:draft');
    v_create_operation := analytics_catalog_uuid_v7_v1(
        p_component_id,'query:create-op'
    );
    v_publish_operation := analytics_catalog_uuid_v7_v1(
        p_component_id,'query:publish-op'
    );
    v_query_document := pg_catalog.convert_to((p_plan->'query')::text,'UTF8');
    v_parameter_document := pg_catalog.convert_to(
        (p_plan->'parameters')::text,'UTF8'
    );
    v_output_document := pg_catalog.convert_to((p_plan->'output')::text,'UTF8');
    v_manifest := analytics_catalog_dependencies_manifest_v1(v_dependencies);
    PERFORM 1 FROM analytics_query_create_v1(
        p_scope_schema,v_create_operation,p_actor_id,
        'catalog.analytics.'||p_component_id::text||'.create',
        pg_catalog.sha256(pg_catalog.convert_to(p_plan::text||':create','UTF8')),
        v_query_id,v_published_id,p_original->>'name',v_query_document,
        v_parameter_document,v_output_document,p_occurred_at
    );
    PERFORM 1 FROM analytics_query_publish_v1(
        p_scope_schema,v_publish_operation,p_actor_id,
        'catalog.analytics.'||p_component_id::text||'.publish',
        pg_catalog.sha256(pg_catalog.convert_to(p_plan::text||':publish','UTF8')),
        v_query_id,v_published_id,v_draft_id,1,1,v_query_document,
        v_parameter_document,v_output_document,v_dependencies,v_manifest,
        'none',p_occurred_at
    );
    SELECT latest.version_id, latest.manifest_sha256
    INTO v_read_version, v_read_manifest
    FROM analytics_query_latest_published_v1(
        p_scope_schema,v_query_id
    ) AS latest;
    SELECT current_query.version_id INTO v_read_draft
    FROM analytics_query_get_v1(p_scope_schema,v_query_id) AS current_query;
    SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'owner_domain',dependency.owner_domain,
        'resource_kind',dependency.resource_kind,
        'resource_id',dependency.resource_id,
        'resource_version_id',dependency.resource_version_id,
        'owner_generation',dependency.owner_generation,
        'catalog_version',dependency.catalog_version,
        'query_contract_version',dependency.query_contract_version,
        'policy_version',dependency.policy_version,
        'data_revision_capability',dependency.data_revision_capability,
        'dependency_sha256',pg_catalog.encode(
            dependency.dependency_sha256,'hex'
        )
    ) ORDER BY dependency.ordinality) INTO v_read_dependencies
    FROM analytics_query_dependencies_v1(
        p_scope_schema,v_query_id,v_published_id
    ) WITH ORDINALITY AS dependency(
        owner_domain,resource_kind,resource_id,resource_version_id,
        owner_generation,catalog_version,query_contract_version,policy_version,
        data_revision_capability,dependency_sha256,ordinality
    );
    IF v_read_version <> v_published_id OR v_read_draft <> v_draft_id OR
       v_read_manifest <> v_manifest OR
       v_read_dependencies IS DISTINCT FROM v_dependencies THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog endpoint operational readback differs';
    END IF;
    RETURN QUERY SELECT 'ready',NULL::text,
        v_query_id,v_published_id,v_draft_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION analytics_catalog_materialize_endpoint_v1(
    name,uuid,uuid,jsonb,jsonb,timestamptz
) FROM PUBLIC;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_endpoint_readiness_v2(
    p_scope_schema name,p_projection_receipt_id uuid
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_projection analytics_catalog_version_projections%ROWTYPE;
    v_plan jsonb;
    v_current jsonb;
    v_stored jsonb;
    v_manifest bytea;
    v_stored_manifest bytea;
    v_count bigint;
BEGIN
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT * INTO v_projection
    FROM analytics_catalog_version_projections AS projection
    WHERE projection.receipt_id=p_projection_receipt_id;
    IF NOT FOUND OR v_projection.analytics_kind<>'endpoint' OR
       v_projection.operational_state<>'ready' OR
       v_projection.ready_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog endpoint readiness evidence differs';
    END IF;
    v_plan:=pg_catalog.convert_from(
        v_projection.runtime_plan_bytes,'UTF8'
    )::jsonb;
    IF v_plan->>'contract'
           IS DISTINCT FROM 'analytics.catalog-runtime-plan.v1' OR
       v_plan->>'kind' IS DISTINCT FROM 'endpoint' OR
       v_plan->>'status' IS DISTINCT FROM 'candidate' OR
       pg_catalog.jsonb_typeof(v_plan->'sources') IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog endpoint pinned plan differs';
    END IF;
    SELECT pg_catalog.jsonb_agg(pg_catalog.jsonb_build_object(
        'owner_domain','iam','resource_kind','form',
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
    ) ORDER BY source.ordinality),pg_catalog.count(*)
    INTO v_current,v_count
    FROM pg_catalog.jsonb_array_elements_text(v_plan->'sources')
        WITH ORDINALITY AS source(value,ordinality)
    JOIN forms_catalog_analytics_dependency_v1(
        p_scope_schema,source.value
    ) AS dependency ON TRUE;
    IF v_count<>pg_catalog.jsonb_array_length(v_plan->'sources') THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog endpoint dependency is unavailable';
    END IF;
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
    ) ORDER BY dependency.ordinal),
    version_row.manifest_sha256
    INTO v_stored,v_stored_manifest
    FROM analytics_saved_queries AS root
    JOIN analytics_saved_query_versions AS version_row
      ON version_row.query_id=root.id
     AND version_row.id=v_projection.published_version_id
    JOIN analytics_query_dependencies AS dependency
      ON dependency.query_version_id=version_row.id
    WHERE root.id=v_projection.operational_resource_id
      AND root.current_draft_version_id=
          v_projection.current_draft_version_id
      AND root.lifecycle_state IN ('staged','retired')
      AND version_row.version_state='published'
    GROUP BY version_row.manifest_sha256;
    v_manifest:=analytics_catalog_dependencies_manifest_v1(v_current);
    IF v_stored IS DISTINCT FROM v_current OR
       v_stored_manifest IS DISTINCT FROM v_manifest THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog endpoint pinned dependency drifted';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_endpoint_readiness_v2(
    name,uuid
) FROM PUBLIC;

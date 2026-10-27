-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_dashboard_readiness_v2(
    p_scope_schema name,p_projection_receipt_id uuid
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_projection analytics_catalog_version_projections%ROWTYPE;
    v_plan jsonb;
    v_manifest_material bytea;
    v_manifest bytea;
    v_stored_manifest bytea;
    v_query_version_id uuid;
    v_query_manifest bytea;
    v_count bigint;
BEGIN
    v_manifest_material:=''::bytea;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT * INTO v_projection
    FROM analytics_catalog_version_projections AS projection
    WHERE projection.receipt_id=p_projection_receipt_id;
    IF NOT FOUND OR v_projection.analytics_kind<>'dashboard' OR
       v_projection.operational_state<>'ready' OR
       v_projection.ready_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog dashboard readiness evidence differs';
    END IF;
    v_plan:=pg_catalog.convert_from(
        v_projection.runtime_plan_bytes,'UTF8'
    )::jsonb;
    IF v_plan->>'contract'
           IS DISTINCT FROM 'analytics.catalog-runtime-plan.v1' OR
       v_plan->>'kind' IS DISTINCT FROM 'dashboard' OR
       v_plan->>'status' IS DISTINCT FROM 'candidate' OR
       pg_catalog.jsonb_typeof(v_plan#>'{dashboard,widgets}')
           IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog dashboard pinned plan differs';
    END IF;
    SELECT version_row.manifest_sha256,pg_catalog.count(widget.widget_id)
    INTO v_stored_manifest,v_count
    FROM analytics_dashboards AS root
    JOIN analytics_dashboard_versions AS version_row
      ON version_row.dashboard_id=root.id
     AND version_row.id=v_projection.published_version_id
    JOIN analytics_dashboard_widgets AS widget
      ON widget.dashboard_version_id=version_row.id
    WHERE root.id=v_projection.operational_resource_id
      AND root.current_draft_version_id=
          v_projection.current_draft_version_id
      AND root.lifecycle_state IN ('staged','retired')
      AND version_row.version_state='published'
    GROUP BY version_row.manifest_sha256;
    IF v_count IS DISTINCT FROM pg_catalog.jsonb_array_length(
        v_plan#>'{dashboard,widgets}'
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog dashboard widget evidence differs';
    END IF;
    FOR v_query_version_id IN
        SELECT widget.query_version_id
        FROM analytics_dashboard_widgets AS widget
        WHERE widget.dashboard_version_id=v_projection.published_version_id
        ORDER BY widget.position
    LOOP
        v_query_manifest:=analytics_catalog_resource_manifest_v2(
            p_scope_schema,'query',v_query_version_id
        );
        v_manifest_material:=v_manifest_material||
            pg_catalog.uuid_send(v_query_version_id)||v_query_manifest;
    END LOOP;
    v_manifest:=pg_catalog.sha256(v_manifest_material);
    IF v_stored_manifest IS DISTINCT FROM v_manifest THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog dashboard manifest drifted';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_dashboard_readiness_v2(
    name,uuid
) FROM PUBLIC;

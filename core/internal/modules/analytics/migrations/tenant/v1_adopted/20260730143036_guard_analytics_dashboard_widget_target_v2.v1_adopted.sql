-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_guard_dashboard_widget_target_v2()
RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_query_id uuid;
    v_root_state text;
    v_component_id uuid;
    v_evidence_count bigint;
    v_owner_generation bigint;
BEGIN
    IF TG_OP='DELETE' THEN
        RETURN OLD;
    END IF;
    SELECT version_row.query_id INTO v_query_id
    FROM analytics_saved_query_versions AS version_row
    WHERE version_row.id=NEW.query_version_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard widget target is unavailable';
    END IF;
    SELECT root.lifecycle_state INTO v_root_state
    FROM analytics_saved_queries AS root
    WHERE root.id=v_query_id;
    IF NOT FOUND OR v_root_state NOT IN ('active','staged') THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard widget target root is inactive';
    END IF;
    IF v_root_state='staged' THEN
        SELECT pg_catalog.count(*),
            pg_catalog.min(candidate.component_id::text)::uuid,
            pg_catalog.min(candidate.owner_generation)
        INTO v_evidence_count,v_component_id,v_owner_generation
        FROM analytics_catalog_component_candidates AS candidate
        JOIN analytics_catalog_version_projections AS projection
          ON projection.receipt_id=candidate.projection_receipt_id
         AND projection.component_id=candidate.component_id
         AND projection.package_id=candidate.package_id
         AND projection.package_version=candidate.package_version
         AND projection.release_id=candidate.release_id
         AND projection.owner_generation=candidate.owner_generation
         AND projection.content_sha256=candidate.content_sha256
        JOIN analytics_catalog_projection_attempts AS attempt
          ON attempt.receipt_id=candidate.attempt_receipt_id
         AND attempt.projection_receipt_id=projection.receipt_id
        WHERE projection.analytics_kind='endpoint'
          AND projection.operational_resource_id=v_query_id
          AND projection.published_version_id=NEW.query_version_id
          AND projection.operational_state='ready'
          AND projection.ready_at IS NOT NULL
          AND attempt.ready_at IS NOT NULL;
        IF v_evidence_count<>1 THEN
            RAISE EXCEPTION USING ERRCODE='P6102',
                MESSAGE='analytics staged widget evidence is ambiguous';
        END IF;
        PERFORM 1 FROM analytics_catalog_components AS component
        WHERE component.component_id=v_component_id
          AND component.owner_generation=v_owner_generation
          AND component.writer_state='v2'
          AND component.component_kind='ENDPOINT'
          AND component.lifecycle_state<>'PURGED'
        FOR SHARE;
        IF NOT FOUND THEN
            RAISE EXCEPTION USING ERRCODE='P6102',
                MESSAGE='analytics staged widget component is unavailable';
        END IF;
        PERFORM 1
        FROM analytics_catalog_component_candidates AS candidate
        JOIN analytics_catalog_version_projections AS projection
          ON projection.receipt_id=candidate.projection_receipt_id
         AND projection.component_id=candidate.component_id
         AND projection.package_id=candidate.package_id
         AND projection.package_version=candidate.package_version
         AND projection.release_id=candidate.release_id
         AND projection.owner_generation=candidate.owner_generation
         AND projection.content_sha256=candidate.content_sha256
        JOIN analytics_catalog_projection_attempts AS attempt
          ON attempt.receipt_id=candidate.attempt_receipt_id
         AND attempt.projection_receipt_id=projection.receipt_id
        WHERE candidate.component_id=v_component_id
          AND projection.analytics_kind='endpoint'
          AND projection.operational_resource_id=v_query_id
          AND projection.published_version_id=NEW.query_version_id
          AND projection.operational_state='ready'
          AND projection.ready_at IS NOT NULL
          AND attempt.ready_at IS NOT NULL
        FOR SHARE OF candidate,projection,attempt;
        GET DIAGNOSTICS v_evidence_count=ROW_COUNT;
        IF v_evidence_count<>1 THEN
            RAISE EXCEPTION USING ERRCODE='P6102',
                MESSAGE='analytics staged widget target lacks Catalog evidence';
        END IF;
    END IF;
    PERFORM 1 FROM analytics_saved_queries AS root
    WHERE root.id=v_query_id AND root.lifecycle_state=v_root_state
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard widget target root changed';
    END IF;
    PERFORM 1
    FROM analytics_saved_query_versions AS version_row
    WHERE version_row.id=NEW.query_version_id
      AND version_row.query_id=v_query_id
      AND (
        (v_root_state='active' AND version_row.version_state IN (
            'published','needs_revalidation'
        )) OR
        (v_root_state='staged' AND version_row.version_state='published')
      )
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='analytics dashboard widget target version is inactive';
    END IF;
    RETURN NEW;
END;
$function$;

REVOKE ALL ON FUNCTION analytics_guard_dashboard_widget_target_v2()
    FROM PUBLIC;

CREATE TRIGGER analytics_dashboard_widgets_target_guard
BEFORE INSERT OR UPDATE OF query_version_id
ON analytics_dashboard_widgets
FOR EACH ROW EXECUTE FUNCTION analytics_guard_dashboard_widget_target_v2();
ALTER TABLE analytics_dashboard_widgets
    ENABLE ALWAYS TRIGGER analytics_dashboard_widgets_target_guard;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
UPDATE analytics_catalog_components AS component
SET lifecycle_state='RETIRED'
FROM analytics_catalog_component_heads AS head
JOIN analytics_catalog_version_projections AS projection
  ON projection.receipt_id=head.projection_receipt_id
 AND projection.component_id=head.component_id
 AND projection.content_sha256=head.content_sha256
WHERE component.component_id=head.component_id
  AND component.lifecycle_state='PURGED'
  AND component.component_kind='ENDPOINT'
  AND component.content_sha256=head.content_sha256
  AND projection.analytics_kind='endpoint'
  AND projection.operational_state='ready'
  AND EXISTS (
      SELECT 1
      FROM analytics_saved_queries AS root
      JOIN analytics_saved_query_versions AS version_row
        ON version_row.query_id=root.id
       AND version_row.id=projection.published_version_id
      WHERE root.id=projection.operational_resource_id
        AND root.lifecycle_state='retired'
  );

UPDATE analytics_catalog_components AS component
SET lifecycle_state='RETIRED'
FROM analytics_catalog_component_heads AS head
JOIN analytics_catalog_version_projections AS projection
  ON projection.receipt_id=head.projection_receipt_id
 AND projection.component_id=head.component_id
 AND projection.content_sha256=head.content_sha256
WHERE component.component_id=head.component_id
  AND component.lifecycle_state='PURGED'
  AND component.component_kind='DASHBOARD'
  AND component.content_sha256=head.content_sha256
  AND projection.analytics_kind='dashboard'
  AND projection.operational_state='ready'
  AND EXISTS (
      SELECT 1
      FROM analytics_dashboards AS root
      JOIN analytics_dashboard_versions AS version_row
        ON version_row.dashboard_id=root.id
       AND version_row.id=projection.published_version_id
      WHERE root.id=projection.operational_resource_id
        AND root.lifecycle_state='retired'
  );

CREATE FUNCTION analytics_catalog_internal_widget_peer_v2(
    p_package_id uuid,p_package_version text,p_release_id uuid,
    p_dashboard_version_id uuid
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_count bigint;
BEGIN
    IF pg_catalog.uuid_extract_version(p_package_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_release_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_dashboard_version_id)
           IS DISTINCT FROM 7 OR p_package_version IS NULL THEN
        RETURN false;
    END IF;
    PERFORM 1
    FROM analytics_catalog_component_heads AS peer_head
    JOIN analytics_catalog_components AS peer_component
      ON peer_component.component_id=peer_head.component_id
     AND peer_component.content_sha256=peer_head.content_sha256
    JOIN analytics_catalog_version_projections AS peer_projection
      ON peer_projection.receipt_id=peer_head.projection_receipt_id
     AND peer_projection.component_id=peer_head.component_id
     AND peer_projection.package_id=peer_head.package_id
     AND peer_projection.package_version=peer_head.package_version
     AND peer_projection.release_id=peer_head.release_id
     AND peer_projection.content_sha256=peer_head.content_sha256
    JOIN analytics_catalog_projection_activations AS activation
      ON activation.projection_receipt_id=peer_head.projection_receipt_id
     AND activation.component_id=peer_head.component_id
     AND activation.package_id=peer_head.package_id
     AND activation.package_version=peer_head.package_version
     AND activation.release_id=peer_head.release_id
     AND activation.content_sha256=peer_head.content_sha256
     AND activation.activated_at=peer_head.activated_at
     AND activation.superseded_at IS NULL
    JOIN analytics_catalog_projection_attempts AS attempt
      ON attempt.receipt_id=activation.attempt_receipt_id
     AND attempt.projection_receipt_id=peer_projection.receipt_id
    JOIN analytics_catalog_definitions AS definition
      ON definition.component_id=peer_projection.component_id
     AND definition.package_id=peer_projection.package_id
     AND definition.package_version=peer_projection.package_version
     AND definition.release_id=peer_projection.release_id
     AND definition.content_sha256=peer_projection.content_sha256
     AND definition.operational_resource_id=
         peer_projection.operational_resource_id
     AND definition.published_version_id=
         peer_projection.published_version_id
     AND definition.current_draft_version_id=
         peer_projection.current_draft_version_id
    JOIN analytics_dashboards AS root
      ON root.id=peer_projection.operational_resource_id
     AND root.lifecycle_state='retired'
    JOIN analytics_dashboard_versions AS owner_version
      ON owner_version.id=p_dashboard_version_id
     AND owner_version.dashboard_id=root.id
    WHERE peer_head.package_id=p_package_id
      AND peer_head.package_version=p_package_version
      AND peer_head.release_id=p_release_id
      AND peer_component.component_kind='DASHBOARD'
      AND peer_component.writer_state='v2'
      AND peer_component.lifecycle_state='RETIRED'
      AND peer_component.owner_generation=peer_projection.owner_generation
      AND peer_component.contract=peer_projection.contract
      AND peer_projection.analytics_kind='dashboard'
      AND peer_projection.component_kind='DASHBOARD'
      AND peer_projection.contract='analytics.catalog-definition.v1'
      AND peer_projection.operational_state='ready'
      AND peer_projection.ready_at IS NOT NULL
      AND attempt.ready_at IS NOT NULL
      AND p_dashboard_version_id IN (
          peer_projection.published_version_id,
          peer_projection.current_draft_version_id
      )
    FOR SHARE OF peer_component,peer_head,peer_projection,activation,
        attempt,definition,root,owner_version;
    GET DIAGNOSTICS v_count=ROW_COUNT;
    RETURN v_count=1;
END;
$function$;
REVOKE ALL ON FUNCTION analytics_catalog_internal_widget_peer_v2(
    uuid,text,uuid,uuid
) FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION analytics_catalog_internal_widget_peer_v2(
    uuid,text,uuid,uuid
) FROM harpy_runtime;
-- +goose StatementEnd

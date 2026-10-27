-- +goose Up
ALTER TABLE analytics_saved_queries
    DROP CONSTRAINT analytics_saved_queries_state_check;
ALTER TABLE analytics_saved_queries
    ADD CONSTRAINT analytics_saved_queries_state_check CHECK (
        lifecycle_state IN ('active','staged','retired')
        AND row_revision>0 AND reference_epoch>0
    );
ALTER TABLE analytics_saved_queries
    DROP CONSTRAINT analytics_saved_queries_draft_presence_check;
ALTER TABLE analytics_saved_queries
    ADD CONSTRAINT analytics_saved_queries_draft_presence_check CHECK (
        (lifecycle_state IN ('active','staged')
         AND current_draft_version_id IS NOT NULL)
        OR lifecycle_state='retired'
    );

ALTER TABLE analytics_dashboards
    DROP CONSTRAINT analytics_dashboards_state_check;
ALTER TABLE analytics_dashboards
    ADD CONSTRAINT analytics_dashboards_state_check CHECK (
        lifecycle_state IN ('active','staged','retired')
        AND row_revision>0 AND reference_epoch>0
    );
ALTER TABLE analytics_dashboards
    DROP CONSTRAINT analytics_dashboards_draft_presence_check;
ALTER TABLE analytics_dashboards
    ADD CONSTRAINT analytics_dashboards_draft_presence_check CHECK (
        (lifecycle_state IN ('active','staged')
         AND current_draft_version_id IS NOT NULL)
        OR lifecycle_state='retired'
    );

UPDATE analytics_saved_queries AS root
SET lifecycle_state='staged',row_revision=root.row_revision+1,
    reference_epoch=root.reference_epoch+1,updated_by=attempt.actor_id,
    updated_at=GREATEST(root.updated_at,candidate.prepared_at)
FROM analytics_catalog_component_candidates AS candidate
JOIN analytics_catalog_version_projections AS projection
  ON projection.receipt_id=candidate.projection_receipt_id
JOIN analytics_catalog_projection_attempts AS attempt
  ON attempt.receipt_id=candidate.attempt_receipt_id
 AND attempt.projection_receipt_id=projection.receipt_id
WHERE projection.analytics_kind='endpoint'
  AND root.id=projection.operational_resource_id
  AND root.lifecycle_state='active';

UPDATE analytics_dashboards AS root
SET lifecycle_state='staged',row_revision=root.row_revision+1,
    reference_epoch=root.reference_epoch+1,updated_by=attempt.actor_id,
    updated_at=GREATEST(root.updated_at,candidate.prepared_at)
FROM analytics_catalog_component_candidates AS candidate
JOIN analytics_catalog_version_projections AS projection
  ON projection.receipt_id=candidate.projection_receipt_id
JOIN analytics_catalog_projection_attempts AS attempt
  ON attempt.receipt_id=candidate.attempt_receipt_id
 AND attempt.projection_receipt_id=projection.receipt_id
WHERE projection.analytics_kind='dashboard'
  AND root.id=projection.operational_resource_id
  AND root.lifecycle_state='active';

-- +goose Up
UPDATE analytics_saved_queries AS root
SET lifecycle_state='retired',
    row_revision=root.row_revision+1,
    reference_epoch=root.reference_epoch+1,
    updated_at=GREATEST(root.updated_at,component.updated_at)
FROM analytics_catalog_component_heads AS head
JOIN analytics_catalog_version_projections AS projection
  ON projection.receipt_id=head.projection_receipt_id
JOIN analytics_catalog_components AS component
  ON component.component_id=head.component_id
WHERE root.id=projection.operational_resource_id
  AND projection.analytics_kind='endpoint'
  AND component.lifecycle_state IN ('DISABLED','RETIRED','PURGED')
  AND root.lifecycle_state='active';

UPDATE analytics_dashboards AS root
SET lifecycle_state='retired',
    row_revision=root.row_revision+1,
    reference_epoch=root.reference_epoch+1,
    updated_at=GREATEST(root.updated_at,component.updated_at)
FROM analytics_catalog_component_heads AS head
JOIN analytics_catalog_version_projections AS projection
  ON projection.receipt_id=head.projection_receipt_id
JOIN analytics_catalog_components AS component
  ON component.component_id=head.component_id
WHERE root.id=projection.operational_resource_id
  AND projection.analytics_kind='dashboard'
  AND component.lifecycle_state IN ('DISABLED','RETIRED','PURGED')
  AND root.lifecycle_state='active';

-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_operational_head_visibility_v2(
    p_scope_schema name,p_component_id uuid,p_visible boolean,
    p_completed_at timestamptz
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_projection analytics_catalog_version_projections%ROWTYPE;
    v_attempt analytics_catalog_projection_attempts%ROWTYPE;
    v_state text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_visible IS NULL OR p_completed_at IS NULL OR
       NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog head visibility input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT attempt.* INTO v_attempt
    FROM analytics_catalog_component_heads AS head
    JOIN analytics_catalog_projection_activations AS activation
      ON activation.projection_receipt_id=head.projection_receipt_id
     AND activation.component_id=head.component_id
     AND activation.package_id=head.package_id
     AND activation.package_version=head.package_version
     AND activation.release_id=head.release_id
     AND activation.content_sha256=head.content_sha256
     AND activation.activated_at=head.activated_at
     AND activation.superseded_at IS NULL
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=activation.projection_receipt_id
    JOIN analytics_catalog_projection_attempts AS attempt
      ON attempt.receipt_id=activation.attempt_receipt_id
     AND attempt.projection_receipt_id=projection.receipt_id
    WHERE head.component_id=p_component_id
    FOR UPDATE OF head,activation,projection,attempt;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog active head is unavailable';
    END IF;
    SELECT * INTO v_projection
    FROM analytics_catalog_version_projections AS projection
    WHERE projection.receipt_id=v_attempt.projection_receipt_id;
    IF NOT FOUND OR v_projection.operational_state<>'ready' OR
       v_projection.ready_at IS NULL OR v_attempt.ready_at IS NULL OR
       p_completed_at<v_attempt.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog live head evidence differs';
    END IF;
    IF v_projection.analytics_kind='endpoint' THEN
        SELECT root.lifecycle_state INTO v_state
        FROM analytics_saved_queries AS root
        WHERE root.id=v_projection.operational_resource_id FOR UPDATE;
        IF p_visible AND v_state='retired' THEN
            PERFORM analytics_catalog_endpoint_readiness_v2(
                p_scope_schema,v_projection.receipt_id
            );
            UPDATE analytics_saved_queries
            SET lifecycle_state='active',row_revision=row_revision+1,
                reference_epoch=reference_epoch+1,
                updated_by=v_attempt.actor_id,updated_at=p_completed_at
            WHERE id=v_projection.operational_resource_id
              AND lifecycle_state='retired';
        ELSIF NOT p_visible AND v_state='active' THEN
            UPDATE analytics_saved_queries
            SET lifecycle_state='retired',row_revision=row_revision+1,
                reference_epoch=reference_epoch+1,
                updated_by=v_attempt.actor_id,updated_at=p_completed_at
            WHERE id=v_projection.operational_resource_id
              AND lifecycle_state='active';
        ELSIF (p_visible AND v_state<>'active') OR
              (NOT p_visible AND v_state<>'retired') THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog query visibility differs';
        END IF;
    ELSIF v_projection.analytics_kind='dashboard' THEN
        SELECT root.lifecycle_state INTO v_state
        FROM analytics_dashboards AS root
        WHERE root.id=v_projection.operational_resource_id FOR UPDATE;
        IF p_visible AND v_state='retired' THEN
            PERFORM analytics_catalog_dashboard_readiness_v2(
                p_scope_schema,v_projection.receipt_id
            );
            UPDATE analytics_dashboards
            SET lifecycle_state='active',row_revision=row_revision+1,
                reference_epoch=reference_epoch+1,
                updated_by=v_attempt.actor_id,updated_at=p_completed_at
            WHERE id=v_projection.operational_resource_id
              AND lifecycle_state='retired';
        ELSIF NOT p_visible AND v_state='active' THEN
            UPDATE analytics_dashboards
            SET lifecycle_state='retired',row_revision=row_revision+1,
                reference_epoch=reference_epoch+1,
                updated_by=v_attempt.actor_id,updated_at=p_completed_at
            WHERE id=v_projection.operational_resource_id
              AND lifecycle_state='active';
        ELSIF (p_visible AND v_state<>'active') OR
              (NOT p_visible AND v_state<>'retired') THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog dashboard visibility differs';
        END IF;
    ELSE
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog head kind differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_operational_head_visibility_v2(
    name,uuid,boolean,timestamptz
) FROM PUBLIC;

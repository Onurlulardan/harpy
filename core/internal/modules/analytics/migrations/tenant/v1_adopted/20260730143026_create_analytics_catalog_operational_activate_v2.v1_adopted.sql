-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_operational_activate_v2(
    p_scope_schema name,p_projection_receipt_id uuid,p_operation_id uuid,
    p_visible boolean,p_completed_at timestamptz
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_projection analytics_catalog_version_projections%ROWTYPE;
    v_attempt analytics_catalog_projection_attempts%ROWTYPE;
    v_old_projection analytics_catalog_version_projections%ROWTYPE;
    v_old_resource_id uuid;
    v_state text;
    v_old_state text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_projection_receipt_id)
           IS DISTINCT FROM 7 OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       p_visible IS NULL OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog operational activation input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT attempt.* INTO v_attempt
    FROM analytics_catalog_component_candidates AS candidate
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=candidate.projection_receipt_id
    JOIN analytics_catalog_projection_attempts AS attempt
      ON attempt.receipt_id=candidate.attempt_receipt_id
     AND attempt.projection_receipt_id=projection.receipt_id
    WHERE candidate.projection_receipt_id=p_projection_receipt_id
      AND candidate.staged_operation_id=p_operation_id
    FOR UPDATE OF candidate,projection,attempt;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog operational candidate differs';
    END IF;
    SELECT * INTO v_projection
    FROM analytics_catalog_version_projections AS projection
    WHERE projection.receipt_id=v_attempt.projection_receipt_id;
    IF NOT FOUND OR v_projection.operational_state<>'ready' OR
       v_projection.ready_at IS NULL OR
       v_attempt.ready_at IS NULL OR p_completed_at<v_attempt.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog operational candidate differs';
    END IF;
    SELECT previous.* INTO v_old_projection
    FROM analytics_catalog_component_heads AS head
    JOIN analytics_catalog_version_projections AS previous
      ON previous.receipt_id=head.projection_receipt_id
    WHERE head.component_id=v_projection.component_id FOR UPDATE OF head;
    IF FOUND THEN
        v_old_resource_id:=v_old_projection.operational_resource_id;
        IF v_old_projection.analytics_kind<>v_projection.analytics_kind THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog operational kind changed';
        END IF;
    END IF;
    IF v_projection.analytics_kind='endpoint' THEN
        PERFORM 1 FROM analytics_saved_queries AS root
        WHERE root.id IN (
            v_projection.operational_resource_id,v_old_resource_id
        ) ORDER BY root.id FOR UPDATE;
        PERFORM analytics_catalog_endpoint_readiness_v2(
            p_scope_schema,p_projection_receipt_id
        );
        SELECT lifecycle_state INTO v_state
        FROM analytics_saved_queries
        WHERE id=v_projection.operational_resource_id;
        IF v_state IS DISTINCT FROM 'staged' THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog candidate query is not staged';
        END IF;
        IF v_old_resource_id IS NOT NULL AND
           v_old_resource_id<>v_projection.operational_resource_id THEN
            SELECT lifecycle_state INTO v_old_state
            FROM analytics_saved_queries
            WHERE id=v_old_resource_id;
            IF NOT FOUND OR (p_visible AND v_old_state<>'active') OR
               (NOT p_visible AND
                v_old_state NOT IN ('active','retired')) THEN
                RAISE EXCEPTION USING ERRCODE='55000',
                    MESSAGE='Analytics Catalog previous query visibility differs';
            END IF;
            UPDATE analytics_saved_queries
            SET lifecycle_state='retired',row_revision=row_revision+1,
                reference_epoch=reference_epoch+1,
                updated_by=v_attempt.actor_id,updated_at=p_completed_at
            WHERE id=v_old_resource_id AND lifecycle_state='active';
        END IF;
        UPDATE analytics_saved_queries
        SET lifecycle_state=CASE WHEN p_visible THEN 'active' ELSE 'retired' END,
            row_revision=row_revision+1,
            reference_epoch=reference_epoch+1,
            updated_by=v_attempt.actor_id,updated_at=p_completed_at
        WHERE id=v_projection.operational_resource_id
          AND lifecycle_state='staged';
    ELSIF v_projection.analytics_kind='dashboard' THEN
        PERFORM 1 FROM analytics_dashboards AS root
        WHERE root.id IN (
            v_projection.operational_resource_id,v_old_resource_id
        ) ORDER BY root.id FOR UPDATE;
        PERFORM analytics_catalog_dashboard_readiness_v2(
            p_scope_schema,p_projection_receipt_id
        );
        SELECT lifecycle_state INTO v_state
        FROM analytics_dashboards
        WHERE id=v_projection.operational_resource_id;
        IF v_state IS DISTINCT FROM 'staged' THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog candidate dashboard is not staged';
        END IF;
        IF v_old_resource_id IS NOT NULL AND
           v_old_resource_id<>v_projection.operational_resource_id THEN
            SELECT lifecycle_state INTO v_old_state
            FROM analytics_dashboards
            WHERE id=v_old_resource_id;
            IF NOT FOUND OR (p_visible AND v_old_state<>'active') OR
               (NOT p_visible AND
                v_old_state NOT IN ('active','retired')) THEN
                RAISE EXCEPTION USING ERRCODE='55000',
                    MESSAGE='Analytics Catalog previous dashboard visibility differs';
            END IF;
            UPDATE analytics_dashboards
            SET lifecycle_state='retired',row_revision=row_revision+1,
                reference_epoch=reference_epoch+1,
                updated_by=v_attempt.actor_id,updated_at=p_completed_at
            WHERE id=v_old_resource_id AND lifecycle_state='active';
        END IF;
        UPDATE analytics_dashboards
        SET lifecycle_state=CASE WHEN p_visible THEN 'active' ELSE 'retired' END,
            row_revision=row_revision+1,
            reference_epoch=reference_epoch+1,
            updated_by=v_attempt.actor_id,updated_at=p_completed_at
        WHERE id=v_projection.operational_resource_id
          AND lifecycle_state='staged';
    ELSE
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog operational kind differs';
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog operational activation was not applied';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_operational_activate_v2(
    name,uuid,uuid,boolean,timestamptz
) FROM PUBLIC;

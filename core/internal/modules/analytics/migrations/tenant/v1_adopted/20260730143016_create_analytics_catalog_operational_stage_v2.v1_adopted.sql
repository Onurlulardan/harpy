-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_operational_stage_v2(
    p_scope_schema name,p_attempt_receipt_id uuid,p_occurred_at timestamptz
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_attempt analytics_catalog_projection_attempts%ROWTYPE;
    v_projection analytics_catalog_version_projections%ROWTYPE;
    v_root_state text;
    v_draft_id uuid;
    v_version_state text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_attempt_receipt_id)
           IS DISTINCT FROM 7 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) OR
       date_trunc('microseconds',p_occurred_at)<>p_occurred_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog operational stage input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT attempt.* INTO v_attempt
    FROM analytics_catalog_projection_attempts AS attempt
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=attempt.projection_receipt_id
    WHERE attempt.receipt_id=p_attempt_receipt_id
    FOR UPDATE OF attempt,projection;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog operational stage evidence differs';
    END IF;
    SELECT * INTO v_projection
    FROM analytics_catalog_version_projections AS projection
    WHERE projection.receipt_id=v_attempt.projection_receipt_id;
    IF NOT FOUND OR v_projection.operational_state<>'ready' OR
       v_projection.ready_at IS NULL OR v_attempt.ready_at IS NULL OR
       p_occurred_at<v_attempt.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog operational stage evidence differs';
    END IF;
    IF EXISTS (
        SELECT 1 FROM analytics_catalog_projection_activations AS activation
        WHERE activation.projection_receipt_id=v_projection.receipt_id
    ) THEN
        RETURN;
    END IF;
    IF v_projection.analytics_kind='endpoint' THEN
        SELECT root.lifecycle_state,root.current_draft_version_id,
            version_row.version_state
        INTO v_root_state,v_draft_id,v_version_state
        FROM analytics_saved_queries AS root
        JOIN analytics_saved_query_versions AS version_row
          ON version_row.id=v_projection.published_version_id
         AND version_row.query_id=root.id
        WHERE root.id=v_projection.operational_resource_id FOR UPDATE OF root;
        IF NOT FOUND OR v_draft_id<>v_projection.current_draft_version_id OR
           v_version_state<>'published' OR
           v_root_state NOT IN ('active','staged','retired') THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog staged query readback differs';
        END IF;
        IF v_root_state<>'staged' THEN
            UPDATE analytics_saved_queries
            SET lifecycle_state='staged',row_revision=row_revision+1,
                reference_epoch=reference_epoch+1,
                updated_by=v_attempt.actor_id,updated_at=p_occurred_at
            WHERE id=v_projection.operational_resource_id;
        END IF;
    ELSIF v_projection.analytics_kind='dashboard' THEN
        SELECT root.lifecycle_state,root.current_draft_version_id,
            version_row.version_state
        INTO v_root_state,v_draft_id,v_version_state
        FROM analytics_dashboards AS root
        JOIN analytics_dashboard_versions AS version_row
          ON version_row.id=v_projection.published_version_id
         AND version_row.dashboard_id=root.id
        WHERE root.id=v_projection.operational_resource_id FOR UPDATE OF root;
        IF NOT FOUND OR v_draft_id<>v_projection.current_draft_version_id OR
           v_version_state<>'published' OR
           v_root_state NOT IN ('active','staged','retired') THEN
            RAISE EXCEPTION USING ERRCODE='55000',
                MESSAGE='Analytics Catalog staged dashboard readback differs';
        END IF;
        IF v_root_state<>'staged' THEN
            UPDATE analytics_dashboards
            SET lifecycle_state='staged',row_revision=row_revision+1,
                reference_epoch=reference_epoch+1,
                updated_by=v_attempt.actor_id,updated_at=p_occurred_at
            WHERE id=v_projection.operational_resource_id;
        END IF;
    ELSE
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog staged resource kind differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_operational_stage_v2(
    name,uuid,timestamptz
) FROM PUBLIC;

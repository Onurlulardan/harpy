-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION analytics_catalog_component_lifecycle_v2(
    p_scope_schema name,p_operation_id uuid,p_component_id uuid,
    p_generation bigint,p_contract text,p_content_sha256 text,
    p_action text,p_reference_epoch bigint,p_completed_at timestamptz
)
RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_component analytics_catalog_components%ROWTYPE;
    v_next text;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       p_generation IS NULL OR p_generation<1 OR p_contract IS NULL OR
       p_content_sha256 IS NULL OR p_content_sha256!~'^[0-9a-f]{64}$' OR
       p_action IS NULL OR
       p_action NOT IN ('disable','enable','retire','restore','purge') OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog lifecycle input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT * INTO v_component
    FROM analytics_catalog_components AS component
    WHERE component.component_id=p_component_id FOR UPDATE;
    IF NOT FOUND OR
       v_component.component_kind NOT IN ('ENDPOINT','DASHBOARD') OR
       v_component.owner_generation<>p_generation OR
       v_component.writer_state<>'v2' OR
       v_component.contract<>p_contract OR
       v_component.content_sha256<>p_content_sha256 OR
       p_completed_at<v_component.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog lifecycle authority differs';
    END IF;
    v_next:=CASE
      WHEN p_action='disable' AND v_component.lifecycle_state='ACTIVE'
        THEN 'DISABLED'
      WHEN p_action='enable' AND v_component.lifecycle_state='DISABLED'
        THEN 'ACTIVE'
      WHEN p_action='retire' AND
           v_component.lifecycle_state IN ('ACTIVE','DISABLED')
        THEN 'RETIRED'
      WHEN p_action='restore' AND v_component.lifecycle_state='RETIRED'
        THEN 'ACTIVE'
      WHEN p_action='purge' AND v_component.lifecycle_state='RETIRED'
        THEN 'PURGED'
      ELSE NULL END;
    IF v_next IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog lifecycle transition is invalid';
    END IF;
    IF p_action='purge' AND (
        p_reference_epoch<1 OR EXISTS (
            SELECT 1 FROM analytics_catalog_component_candidates AS candidate
            WHERE candidate.component_id=p_component_id
        ) OR NOT EXISTS (
            SELECT 1 FROM analytics_catalog_blocker_receipts AS blocker
            WHERE blocker.operation_id=p_operation_id
              AND blocker.component_id=p_component_id
              AND blocker.reference_epoch=p_reference_epoch
              AND blocker.blockers='[]'::jsonb
        )
    ) THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog purge evidence is unavailable';
    END IF;
    IF p_action IN ('disable','retire') THEN
        PERFORM analytics_catalog_operational_head_visibility_v2(
            p_scope_schema,p_component_id,false,p_completed_at
        );
    ELSIF p_action IN ('enable','restore') THEN
        PERFORM analytics_catalog_operational_head_visibility_v2(
            p_scope_schema,p_component_id,true,p_completed_at
        );
    ELSIF p_action='purge' THEN
        PERFORM analytics_catalog_operational_head_visibility_v2(
            p_scope_schema,p_component_id,false,p_completed_at
        );
    END IF;
    UPDATE analytics_catalog_components AS component
    SET lifecycle_state=v_next,updated_at=p_completed_at
    WHERE component.component_id=p_component_id;
    RETURN v_next;
END;
$function$;
REVOKE ALL ON FUNCTION analytics_catalog_component_lifecycle_v2(
    name,uuid,uuid,bigint,text,text,text,bigint,timestamptz
) FROM PUBLIC;

CREATE FUNCTION analytics_catalog_candidate_cleanup_v2(
    p_scope_schema name,p_component_id uuid,p_operation_id uuid,
    p_content_sha256 text,p_completed_at timestamptz
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_projection analytics_catalog_version_projections%ROWTYPE;
    v_attempt analytics_catalog_projection_attempts%ROWTYPE;
BEGIN
    IF pg_catalog.uuid_extract_version(p_component_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       p_content_sha256 IS NULL OR
       p_content_sha256!~'^[0-9a-f]{64}$' OR
       p_completed_at IS NULL OR NOT pg_catalog.isfinite(p_completed_at) OR
       date_trunc('microseconds',p_completed_at)<>p_completed_at THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics Catalog cleanup input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT attempt.* INTO v_attempt
    FROM analytics_catalog_component_candidates AS candidate
    JOIN analytics_catalog_version_projections AS projection
      ON projection.receipt_id=candidate.projection_receipt_id
    JOIN analytics_catalog_projection_attempts AS attempt
      ON attempt.receipt_id=candidate.attempt_receipt_id
     AND attempt.projection_receipt_id=projection.receipt_id
    WHERE candidate.component_id=p_component_id
      AND candidate.content_sha256=p_content_sha256
      AND candidate.staged_operation_id=p_operation_id
    FOR UPDATE OF candidate,projection,attempt;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog cleanup candidate differs';
    END IF;
    SELECT * INTO v_projection
    FROM analytics_catalog_version_projections AS projection
    WHERE projection.receipt_id=v_attempt.projection_receipt_id;
    IF NOT FOUND OR v_projection.operational_state<>'ready' OR
       v_projection.ready_at IS NULL OR v_attempt.ready_at IS NULL OR
       p_completed_at<v_attempt.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog cleanup candidate differs';
    END IF;
    IF v_projection.analytics_kind='endpoint' THEN
        UPDATE analytics_saved_queries
        SET lifecycle_state='retired',row_revision=row_revision+1,
            reference_epoch=reference_epoch+1,
            updated_by=v_attempt.actor_id,updated_at=p_completed_at
        WHERE id=v_projection.operational_resource_id
          AND lifecycle_state='staged';
    ELSIF v_projection.analytics_kind='dashboard' THEN
        UPDATE analytics_dashboards
        SET lifecycle_state='retired',row_revision=row_revision+1,
            reference_epoch=reference_epoch+1,
            updated_by=v_attempt.actor_id,updated_at=p_completed_at
        WHERE id=v_projection.operational_resource_id
          AND lifecycle_state='staged';
    ELSE
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog cleanup resource kind differs';
    END IF;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog staged resource cleanup differs';
    END IF;
    DELETE FROM analytics_catalog_component_candidates AS candidate
    WHERE candidate.component_id=p_component_id
      AND candidate.projection_receipt_id=v_projection.receipt_id
      AND candidate.attempt_receipt_id=v_attempt.receipt_id
      AND candidate.staged_operation_id=p_operation_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='55000',
            MESSAGE='Analytics Catalog candidate cleanup differs';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION analytics_catalog_candidate_cleanup_v2(
    name,uuid,uuid,text,timestamptz
) FROM PUBLIC;

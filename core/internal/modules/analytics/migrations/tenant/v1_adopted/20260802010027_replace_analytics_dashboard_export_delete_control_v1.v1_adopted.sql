-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION analytics_dashboard_export_delete_v1(
    p_scope_schema name,p_operation_id uuid,p_delete_job_id uuid,
    p_actor_id uuid,p_idempotency_key text,p_request_sha256 bytea,
    p_dashboard_id uuid,p_export_id uuid,p_occurred_at timestamptz
)
RETURNS TABLE(
    export_id uuid,dashboard_id uuid,dashboard_version_id uuid,job_id uuid,
    effective_state text,stored_state text,format text,filename text,
    content_type text,failure_code text,row_count bigint,
    plaintext_bytes bigint,cancellation_requested boolean,
    created_at timestamptz,updated_at timestamptz,expires_at timestamptz,
    replayed boolean,schedule_delete boolean
)
LANGUAGE plpgsql SECURITY DEFINER SET search_path FROM CURRENT
AS $function$
DECLARE
    v_export analytics_dashboard_exports%ROWTYPE;
    v_control_state text;
    v_control_changed_at timestamptz;
    v_update_at timestamptz;
    v_schedule boolean:=false;
    v_immediate_delete boolean:=false;
BEGIN
    IF pg_catalog.uuid_extract_version(p_operation_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_delete_job_id) IS DISTINCT FROM 7 OR
       p_actor_id IS NULL OR
       p_actor_id='00000000-0000-0000-0000-000000000000'::uuid OR
       pg_catalog.octet_length(p_idempotency_key) NOT BETWEEN 16 AND 128 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.uuid_extract_version(p_dashboard_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_occurred_at IS NULL OR NOT pg_catalog.isfinite(p_occurred_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export delete input is invalid';
    END IF;
    PERFORM analytics_assert_current_authority_v1(p_scope_schema);
    SELECT item.* INTO v_export FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.dashboard_id=p_dashboard_id;
    IF NOT FOUND OR NOT analytics_dashboard_export_access_v1(
        p_actor_id,p_dashboard_id,v_export.dashboard_version_id) THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='Analytics dashboard export was not found';
    END IF;
    SELECT control.control_state,control.changed_at
    INTO STRICT v_control_state,v_control_changed_at
    FROM platform.platform_async_analytics_dashboard_export_control_v1(
        v_export.tenant_id,p_export_id,v_export.job_id,p_occurred_at
    ) AS control;
    PERFORM pg_catalog.pg_advisory_xact_lock(pg_catalog.hashtextextended(
        'analytics:dashboard:export:delete:'||p_export_id::text,6100));
    SELECT item.* INTO STRICT v_export
    FROM analytics_dashboard_exports AS item
    WHERE item.export_id=p_export_id AND item.dashboard_id=p_dashboard_id
      AND item.tenant_id=v_export.tenant_id AND item.job_id=v_export.job_id
    FOR UPDATE;
    IF NOT analytics_dashboard_export_access_v1(
        p_actor_id,p_dashboard_id,v_export.dashboard_version_id) THEN
        RAISE EXCEPTION USING ERRCODE='P6102',
            MESSAGE='Analytics dashboard export was not found';
    END IF;
    IF v_export.delete_operation_id IS NOT NULL THEN
        IF v_export.expiry_job_id IS NOT NULL AND
           v_export.delete_operation_id=v_export.expiry_job_id AND
           v_export.delete_job_id=v_export.expiry_job_id THEN
            RETURN QUERY SELECT result.*,false,false
            FROM analytics_dashboard_export_snapshot_v1(
                p_scope_schema,p_actor_id,p_dashboard_id,p_export_id,p_occurred_at
            ) AS result;
            RETURN;
        END IF;
        IF v_export.delete_actor_id IS DISTINCT FROM p_actor_id OR
           v_export.delete_idempotency_key IS DISTINCT FROM p_idempotency_key OR
           v_export.delete_request_sha256 IS DISTINCT FROM p_request_sha256 THEN
            RAISE EXCEPTION USING ERRCODE='P6103',
                MESSAGE='Analytics dashboard export delete replay differs';
        END IF;
        RETURN QUERY SELECT result.*,true,false
        FROM analytics_dashboard_export_snapshot_v1(
            p_scope_schema,p_actor_id,p_dashboard_id,p_export_id,p_occurred_at
        ) AS result;
        RETURN;
    END IF;
    IF v_control_state='cancelled' THEN
        IF v_export.state NOT IN ('queued','cancelled') OR
           v_export.quota_state NOT IN ('none','released') THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export cancellation differs';
        END IF;
    ELSIF v_control_state='cancellation_requested' THEN
        IF v_export.state IS DISTINCT FROM 'running' OR
           v_export.quota_state IS DISTINCT FROM 'held' OR
           v_export.platform_attempt_id IS NULL THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export cancellation request differs';
        END IF;
    ELSIF v_control_state='outcome_unknown' THEN
        IF v_export.state IS DISTINCT FROM 'running' OR
           v_export.quota_state IS DISTINCT FROM 'held' OR
           v_export.failure_code IS DISTINCT FROM 'ARTIFACT_OUTCOME_UNKNOWN' OR
           v_export.lease_owner IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export unknown control differs';
        END IF;
    ELSIF v_control_state='succeeded' THEN
        IF v_export.state NOT IN ('ready','deleting','deleted') OR
           v_export.quota_state NOT IN ('used','released') THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export success control differs';
        END IF;
        v_schedule:=v_export.state='ready' AND v_export.quota_state='used';
    ELSIF v_control_state IN ('dead','unknown_resolved') THEN
        IF v_export.state NOT IN ('failed','cancelled','expired','deleted') OR
           v_export.quota_state IS DISTINCT FROM 'released' THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export terminal control differs';
        END IF;
        v_immediate_delete:=v_export.state IS DISTINCT FROM 'deleted';
    ELSE
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export control state is invalid';
    END IF;
    v_update_at:=GREATEST(
        p_occurred_at,v_control_changed_at,
        v_export.updated_at+interval '1 microsecond'
    );
    UPDATE analytics_dashboard_exports
    SET state=CASE
            WHEN v_control_state='cancelled' THEN 'cancelled'
            WHEN v_schedule THEN 'deleting'
            WHEN v_immediate_delete THEN 'deleted'
            ELSE state END,
        quota_state=CASE
            WHEN v_control_state='cancelled' OR v_immediate_delete
                THEN 'released' ELSE quota_state END,
        cancellation_requested_at=CASE
            WHEN v_control_state IN ('cancellation_requested','outcome_unknown')
                THEN COALESCE(cancellation_requested_at,p_occurred_at)
            ELSE cancellation_requested_at END,
        delete_operation_id=p_operation_id,delete_job_id=p_delete_job_id,
        delete_actor_id=p_actor_id,delete_idempotency_key=p_idempotency_key,
        delete_request_sha256=p_request_sha256,
        deleted_at=CASE WHEN v_immediate_delete THEN v_update_at ELSE deleted_at END,
        row_revision=row_revision+1,updated_at=v_update_at
    WHERE analytics_dashboard_exports.export_id=p_export_id;
    RETURN QUERY SELECT result.*,false,v_schedule
    FROM analytics_dashboard_export_snapshot_v1(
        p_scope_schema,p_actor_id,p_dashboard_id,p_export_id,v_update_at
    ) AS result;
END;
$function$;
-- +goose StatementEnd

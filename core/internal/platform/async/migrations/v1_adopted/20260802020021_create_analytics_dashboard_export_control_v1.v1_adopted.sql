-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_control_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,
    p_requested_at timestamptz
)
RETURNS TABLE(control_state text,changed_at timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_resolution platform.analytics_dashboard_export_unknown_resolutions%ROWTYPE;
    v_changed_at timestamptz;
    v_error_digest text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export control authority is invalid';
    END IF;
    IF p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       NOT pg_catalog.isfinite(p_requested_at) OR
       p_requested_at>pg_catalog.clock_timestamp()+interval '5 seconds' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export control input is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_domain='analytics' AND job.writer_origin='v2'
      AND job.kind='analytics.dashboard.export.render'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.max_attempts=3 AND job.payload_ref=p_export_id::text
      AND job.payload_sha256=pg_catalog.encode(
          pg_catalog.sha256(pg_catalog.convert_to(p_export_id::text,'UTF8')),'hex')
      AND job.dedupe_key='analytics.dashboard.export.render:'||p_export_id::text
      AND job.effect_idempotency_key=job.dedupe_key FOR UPDATE;
    IF v_job.id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export control evidence differs';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_domain);
    IF v_job.state IN ('succeeded','dead','cancelled') THEN
        SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
        WHERE attempt.job_id=p_job_id AND attempt.scope_kind='tenant'
          AND attempt.scope_id=p_tenant_id AND attempt.attempt=v_job.attempts
          AND attempt.lease_generation=v_job.lease_generation FOR SHARE;
        IF v_job.completed_at IS NULL OR v_job.completed_at<>v_job.updated_at THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export terminal evidence differs';
        END IF;
        IF v_job.state='dead' AND v_job.last_error_class IS NULL AND
           v_job.last_error_code='JOB_OUTCOME_UNKNOWN' AND
           v_job.last_error_sha256 IS NULL THEN
            IF NOT COALESCE((
               v_attempt.attempt_id IS NOT NULL AND
               v_attempt.finished_at=v_job.completed_at AND
               v_attempt.outcome='dead' AND v_attempt.error_class IS NULL AND
               v_attempt.error_code='JOB_OUTCOME_UNKNOWN' AND
               v_attempt.redacted_error_sha256 IS NULL AND
               v_attempt.effect_started_at IS NOT NULL AND
               v_attempt.effect_receipt_id IS NULL),false) THEN
                RAISE EXCEPTION USING ERRCODE='P6048',
                    MESSAGE='Analytics dashboard export unknown evidence differs';
            END IF;
            SELECT resolution.* INTO v_resolution
            FROM platform.analytics_dashboard_export_unknown_resolutions AS resolution
            WHERE resolution.job_id=p_job_id FOR SHARE;
            IF FOUND AND (v_resolution.tenant_id<>p_tenant_id OR
               v_resolution.export_id<>p_export_id OR
               v_resolution.attempt_id<>v_attempt.attempt_id OR
               v_resolution.attempt_generation<>v_attempt.attempt OR
               v_resolution.lease_owner_id<>v_attempt.lease_owner_id OR
               v_resolution.effect_id<>v_attempt.effect_id OR
               v_resolution.request_sha256<>v_attempt.effect_request_sha256) THEN
                RAISE EXCEPTION USING ERRCODE='P6048',
                    MESSAGE='Analytics dashboard export resolution evidence differs';
            END IF;
            control_state:=CASE WHEN FOUND THEN 'unknown_resolved'
                ELSE 'outcome_unknown' END;
            changed_at:=COALESCE(v_resolution.resolved_at,v_job.completed_at);
            RETURN NEXT; RETURN;
        END IF;
        IF v_job.state='succeeded' AND NOT COALESCE((
           v_attempt.attempt_id IS NOT NULL AND
           v_job.last_error_class IS NULL AND v_job.last_error_code IS NULL AND
           v_job.last_error_sha256 IS NULL AND
           v_attempt.finished_at=v_job.completed_at AND
           v_attempt.outcome='succeeded' AND
           v_attempt.effect_receipt_id IS NOT NULL),false) THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export success evidence differs';
        ELSIF v_job.state='dead' AND NOT COALESCE((
           v_attempt.attempt_id IS NOT NULL AND
           v_job.last_error_class IN ('transient','permanent','conflict') AND
           v_attempt.finished_at=v_job.completed_at AND
           v_attempt.outcome='dead' AND
           v_attempt.error_class=v_job.last_error_class AND
           v_attempt.error_code=v_job.last_error_code AND
            v_attempt.redacted_error_sha256=v_job.last_error_sha256 AND
           ((v_attempt.effect_started_at IS NULL AND
             v_attempt.effect_receipt_id IS NULL) OR
            (v_attempt.effect_started_at IS NOT NULL AND
             v_attempt.effect_receipt_id IS NOT NULL))),false) THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export failure evidence differs';
        ELSIF v_job.state='cancelled' AND NOT COALESCE((
           v_job.last_error_class='cancelled' AND
           v_job.last_error_code='EXPORT_CANCELLED' AND
           v_job.last_error_sha256~'^[0-9a-f]{64}$' AND
           v_job.last_error_sha256<>pg_catalog.repeat('0',64) AND NOT EXISTS (
             SELECT 1 FROM platform.job_attempts AS active
             WHERE active.job_id=p_job_id AND active.finished_at IS NULL
           ) AND ((v_job.attempts=0 AND v_job.lease_generation=0 AND
                   v_attempt.attempt_id IS NULL AND NOT EXISTS (
                     SELECT 1 FROM platform.job_attempts AS unexpected
                     WHERE unexpected.job_id=p_job_id
                   )) OR (v_job.attempts>0 AND
                   v_attempt.attempt_id IS NOT NULL AND
                   v_attempt.finished_at IS NOT NULL AND (
                     (v_attempt.outcome IN ('requeued','lease_expired') AND
                      v_attempt.finished_at<v_job.completed_at) OR
                     (v_attempt.outcome='cancelled' AND
                      v_attempt.finished_at=v_job.completed_at AND
                      v_attempt.error_class=v_job.last_error_class AND
                      v_attempt.error_code=v_job.last_error_code AND
                      v_attempt.redacted_error_sha256=v_job.last_error_sha256)
                   )))),false) THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export cancellation evidence differs';
        END IF;
        control_state:=v_job.state; changed_at:=v_job.completed_at;
        RETURN NEXT; RETURN;
    END IF;
    IF v_job.state='running' AND v_job.cancel_requested_at IS NOT NULL THEN
        control_state:='cancellation_requested';
        changed_at:=v_job.cancel_requested_at;
        RETURN NEXT; RETURN;
    END IF;
    v_changed_at:=GREATEST(pg_catalog.clock_timestamp(),p_requested_at,
        v_job.updated_at+interval '1 microsecond');
    IF v_job.state='queued' AND v_job.effect_started_at IS NULL THEN
        v_error_digest:=pg_catalog.encode(pg_catalog.sha256(
            pg_catalog.convert_to('analytics.dashboard.export.cancel.v1','UTF8')||
            pg_catalog.uuid_send(p_job_id)||pg_catalog.uuid_send(p_export_id)),'hex');
        UPDATE platform.jobs SET state='cancelled',
            last_error_class='cancelled',last_error_code='EXPORT_CANCELLED',
            last_error_sha256=v_error_digest,
            retain_until=v_changed_at+interval '91 days',
            updated_at=v_changed_at,completed_at=v_changed_at
        WHERE id=p_job_id;
        control_state:='cancelled'; changed_at:=v_changed_at;
        RETURN NEXT; RETURN;
    END IF;
    IF v_job.state='running' THEN
        UPDATE platform.jobs SET cancel_requested_at=v_changed_at,
            updated_at=v_changed_at WHERE id=p_job_id;
        control_state:='cancellation_requested'; changed_at:=v_changed_at;
        RETURN NEXT; RETURN;
    END IF;
    RAISE EXCEPTION USING ERRCODE='P6048',
        MESSAGE='Analytics dashboard export control state is invalid';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_control_v1(
        uuid,uuid,uuid,timestamptz)
FROM PUBLIC;

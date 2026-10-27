-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_assert_analytics_dashboard_export_receipt_replay_v1(
    p_receipt platform.idempotency_receipts,p_job platform.jobs,
    p_current_user name,p_session_user name
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_existing platform.idempotency_receipts%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_operation text;
    v_success_code text;
    v_expected_class text;
    v_expected_effect uuid;
    v_expected_receipt uuid;
    v_result_state text;
    v_result_ref text;
BEGIN
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_current_user IS DISTINCT FROM 'forgecrud_owner' OR
       p_session_user IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_receipt.writer_origin IS DISTINCT FROM 'v2' OR
       p_receipt.effect_kind NOT IN ('analytics.dashboard.export.delete',
                                     'analytics.dashboard.export.expire') THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export receipt replay authority is invalid';
    END IF;
    SELECT receipt.* INTO v_existing FROM platform.idempotency_receipts AS receipt
    WHERE receipt.receipt_id=p_receipt.receipt_id FOR SHARE;
    IF NOT FOUND THEN RETURN false; END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.job_id=p_job.id
      AND attempt.effect_receipt_id=p_receipt.receipt_id
      AND attempt.effect_id=p_receipt.effect_id
      AND attempt.effect_request_sha256=p_receipt.request_sha256
    ORDER BY attempt.attempt DESC LIMIT 1 FOR SHARE;
    v_operation:=CASE p_job.kind
        WHEN 'analytics.dashboard.export.delete' THEN 'delete'
        WHEN 'analytics.dashboard.export.expire' THEN 'expire' END;
    v_success_code:=CASE p_job.kind
        WHEN 'analytics.dashboard.export.delete' THEN 'ANALYTICS_EXPORT_DELETED'
        WHEN 'analytics.dashboard.export.expire' THEN
            'ANALYTICS_EXPORT_EXPIRY_COMPLETE' END;
    v_expected_class:=CASE p_receipt.effect_code
        WHEN 'ANALYTICS_EXPORT_REJECTED' THEN 'permanent'
        WHEN 'ANALYTICS_EXPORT_CONFLICT' THEN 'conflict'
        WHEN 'ANALYTICS_EXPORT_UNAVAILABLE' THEN 'transient'
        WHEN 'ANALYTICS_EXPORT_RETRY' THEN 'transient'
        WHEN 'ANALYTICS_EXPORT_TERMINAL' THEN 'permanent' END;
    IF p_job.id IS NULL OR v_attempt.attempt_id IS NULL OR
       v_existing IS DISTINCT FROM p_receipt OR
       pg_catalog.uuid_extract_version(v_attempt.attempt_id) IS DISTINCT FROM 7 OR
       (p_receipt.effect_status='succeeded' AND
          p_receipt.effect_code IS DISTINCT FROM v_success_code) OR
       (p_receipt.effect_status='failed' AND v_expected_class IS NULL) OR
       p_receipt.effect_status NOT IN ('succeeded','failed') THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export receipt replay differs';
    END IF;
    v_expected_effect:=
        platform.platform_async_analytics_export_maintenance_effect_id_v1(
            v_attempt.attempt_id,p_job.kind);
    v_expected_receipt:=
        platform.platform_async_analytics_export_maintenance_receipt_id_v2(
            v_expected_effect,p_receipt.effect_code);
    IF p_receipt.receipt_id IS DISTINCT FROM v_expected_receipt OR
       p_receipt.record_kind IS DISTINCT FROM 'effect' OR
       p_receipt.tenant_id IS DISTINCT FROM p_job.tenant_id OR
       p_receipt.scope_kind IS DISTINCT FROM p_job.scope_kind OR
       p_receipt.scope_id IS DISTINCT FROM p_job.scope_id OR
       p_receipt.authority_group_id IS DISTINCT FROM p_job.authority_group_id OR
       p_receipt.authority_group_generation IS DISTINCT FROM
           p_job.authority_group_generation OR
       p_receipt.authority_domain IS DISTINCT FROM p_job.authority_domain OR
       p_receipt.cutover_generation IS DISTINCT FROM p_job.cutover_generation OR
       p_receipt.retention_class IS DISTINCT FROM 'external_effect' OR
       p_receipt.effect_id IS DISTINCT FROM v_expected_effect OR
       p_receipt.effect_kind IS DISTINCT FROM p_job.kind OR
       p_receipt.effect_revision IS DISTINCT FROM 1 OR
       p_receipt.effect_key IS DISTINCT FROM p_job.effect_idempotency_key OR
       p_receipt.provider IS DISTINCT FROM 'analytics.dashboard-export' OR
       p_receipt.operation IS DISTINCT FROM v_operation OR
       p_receipt.correlation_id IS DISTINCT FROM p_job.id::text OR
       p_receipt.causation_id IS DISTINCT FROM p_job.id OR
       p_receipt.provider_reference IS DISTINCT FROM p_job.payload_ref OR
       p_receipt.retain_until IS DISTINCT FROM
           p_receipt.completed_at+interval '90 days' OR
       v_attempt.finished_at IS NULL OR v_attempt.effect_started_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export receipt replay evidence differs';
    END IF;
    IF p_receipt.effect_status='failed' THEN
        IF v_attempt.outcome NOT IN ('requeued','dead') OR
           v_attempt.error_class IS DISTINCT FROM v_expected_class OR
           v_attempt.error_code IS DISTINCT FROM p_receipt.effect_code OR
           p_receipt.effect_error_class IS DISTINCT FROM v_expected_class OR
           p_receipt.effect_result_ref IS NOT NULL OR
           p_receipt.effect_result_sha256 IS NOT NULL OR
           (v_attempt.outcome='dead' AND p_job.state IS DISTINCT FROM 'dead') THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export failure receipt replay differs';
        END IF;
        RETURN true;
    END IF;
    v_result_state:=pg_catalog.split_part(p_receipt.effect_result_ref,'/',5);
    v_result_ref:='analytics-dashboard-export://'||v_operation||'/'||
        p_job.payload_ref||'/'||v_result_state;
    IF p_job.state IS DISTINCT FROM 'succeeded' OR
       p_job.completed_at IS NULL OR p_job.updated_at IS DISTINCT FROM p_job.completed_at OR
       p_job.lease_owner_id IS NOT NULL OR p_job.lease_until IS NOT NULL OR
       p_job.last_error_class IS NOT NULL OR p_job.last_error_code IS NOT NULL OR
       v_attempt.outcome IS DISTINCT FROM 'succeeded' OR
       v_attempt.finished_at IS DISTINCT FROM p_job.completed_at OR
       v_attempt.error_class IS NOT NULL OR v_attempt.error_code IS NOT NULL OR
       p_receipt.effect_error_class IS NOT NULL OR
       p_receipt.effect_result_ref IS DISTINCT FROM v_result_ref OR
       p_receipt.effect_result_sha256 IS DISTINCT FROM pg_catalog.encode(
          pg_catalog.sha256(pg_catalog.convert_to(v_result_ref,'UTF8')),'hex') OR
       (p_job.kind='analytics.dashboard.export.delete' AND
          v_result_state IS DISTINCT FROM 'deleted') OR
       (p_job.kind='analytics.dashboard.export.expire' AND
          v_result_state NOT IN ('expired','cancelled','failed','deleted')) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export success receipt replay differs';
    END IF;
    RETURN true;
END;
$function$;

CREATE OR REPLACE FUNCTION platform.platform_async_route_analytics_dashboard_export_maintenance_receipt_v1(
    p_receipt platform.idempotency_receipts,p_current_user name,p_session_user name
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_job platform.jobs%ROWTYPE;
BEGIN
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_current_user IS DISTINCT FROM 'forgecrud_owner' OR
       p_session_user IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_receipt.writer_origin IS DISTINCT FROM 'v2' OR
       p_receipt.effect_kind NOT IN ('analytics.dashboard.export.delete',
                                     'analytics.dashboard.export.expire') THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance receipt route is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_receipt.causation_id AND job.tenant_id=p_receipt.tenant_id
    FOR UPDATE;
    IF platform.platform_async_assert_analytics_dashboard_export_receipt_replay_v1(
        p_receipt,v_job,p_current_user,p_session_user) THEN RETURN; END IF;
    IF v_job.state='running' THEN
        PERFORM platform.platform_async_assert_analytics_dashboard_export_running_receipt_v1(
            p_receipt,p_current_user,p_session_user);
    ELSIF v_job.state='dead' AND
          v_job.last_error_code='JOB_OUTCOME_UNKNOWN' THEN
        PERFORM platform.platform_async_assert_analytics_dashboard_export_maintenance_receipt_v1(
            p_receipt,p_current_user,p_session_user);
    ELSE
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance receipt state is invalid';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_assert_analytics_dashboard_export_receipt_replay_v1(
        platform.idempotency_receipts,platform.jobs,name,name) FROM PUBLIC;

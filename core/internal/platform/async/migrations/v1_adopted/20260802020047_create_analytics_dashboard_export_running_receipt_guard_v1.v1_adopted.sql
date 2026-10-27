-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_assert_analytics_dashboard_export_running_receipt_v1(
    p_receipt platform.idempotency_receipts,p_current_user name,p_session_user name
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
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
       p_current_user IS DISTINCT FROM 'harpy_owner' OR
       p_session_user IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_receipt.effect_kind NOT IN ('analytics.dashboard.export.delete',
                                     'analytics.dashboard.export.expire') OR
       p_receipt.writer_origin IS DISTINCT FROM 'v2' OR
       pg_catalog.uuid_extract_version(p_receipt.effect_id) IS DISTINCT FROM 5 OR
       pg_catalog.isfinite(p_receipt.completed_at) IS DISTINCT FROM true THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export running receipt authority is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_receipt.causation_id AND job.tenant_id=p_receipt.tenant_id
    FOR UPDATE;
    IF v_job.id IS NULL OR
       pg_catalog.pg_input_is_valid(v_job.payload_ref,'uuid') IS DISTINCT FROM true THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export running receipt job differs';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_maintenance_effect_owner_v1(
        v_job.tenant_id,v_job.payload_ref::uuid,v_job.id,p_receipt.effect_id,
        p_receipt.completed_at);
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.job_id=v_job.id AND attempt.scope_kind='tenant'
      AND attempt.scope_id=v_job.tenant_id AND attempt.attempt=v_job.attempts
      AND attempt.lease_generation=v_job.lease_generation FOR UPDATE;
    IF pg_catalog.uuid_extract_version(v_attempt.attempt_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export running receipt attempt differs';
    END IF;
    v_operation:=CASE v_job.kind
        WHEN 'analytics.dashboard.export.delete' THEN 'delete'
        WHEN 'analytics.dashboard.export.expire' THEN 'expire' END;
    v_success_code:=CASE v_job.kind
        WHEN 'analytics.dashboard.export.delete' THEN 'ANALYTICS_EXPORT_DELETED'
        WHEN 'analytics.dashboard.export.expire' THEN
            'ANALYTICS_EXPORT_EXPIRY_COMPLETE' END;
    v_expected_class:=CASE p_receipt.effect_code
        WHEN 'ANALYTICS_EXPORT_REJECTED' THEN 'permanent'
        WHEN 'ANALYTICS_EXPORT_CONFLICT' THEN 'conflict'
        WHEN 'ANALYTICS_EXPORT_UNAVAILABLE' THEN 'transient'
        WHEN 'ANALYTICS_EXPORT_RETRY' THEN 'transient'
        WHEN 'ANALYTICS_EXPORT_TERMINAL' THEN 'permanent' END;
    IF (p_receipt.effect_status='succeeded' AND
            p_receipt.effect_code IS DISTINCT FROM v_success_code) OR
       (p_receipt.effect_status='failed' AND v_expected_class IS NULL) OR
       p_receipt.effect_status NOT IN ('succeeded','failed') THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export running receipt outcome differs';
    END IF;
    v_expected_effect:=
        platform.platform_async_analytics_export_maintenance_effect_id_v1(
            v_attempt.attempt_id,v_job.kind);
    v_expected_receipt:=
        platform.platform_async_analytics_export_maintenance_receipt_id_v2(
            v_expected_effect,p_receipt.effect_code);
    IF p_receipt.receipt_id IS DISTINCT FROM v_expected_receipt OR
       p_receipt.record_kind IS DISTINCT FROM 'effect' OR
       p_receipt.scope_kind IS DISTINCT FROM v_job.scope_kind OR
       p_receipt.scope_id IS DISTINCT FROM v_job.scope_id OR
       p_receipt.tenant_id IS DISTINCT FROM v_job.tenant_id OR
       p_receipt.authority_group_id IS DISTINCT FROM v_job.authority_group_id OR
       p_receipt.authority_group_generation IS DISTINCT FROM
           v_job.authority_group_generation OR
       p_receipt.authority_domain IS DISTINCT FROM v_job.authority_domain OR
       p_receipt.cutover_generation IS DISTINCT FROM v_job.cutover_generation OR
       p_receipt.request_sha256 IS DISTINCT FROM v_attempt.effect_request_sha256 OR
       p_receipt.retention_class IS DISTINCT FROM 'external_effect' OR
       p_receipt.effect_id IS DISTINCT FROM v_expected_effect OR
       p_receipt.effect_kind IS DISTINCT FROM v_job.kind OR
       p_receipt.effect_revision IS DISTINCT FROM 1 OR
       p_receipt.effect_key IS DISTINCT FROM v_job.effect_idempotency_key OR
       p_receipt.provider IS DISTINCT FROM 'analytics.dashboard-export' OR
       p_receipt.operation IS DISTINCT FROM v_operation OR
       p_receipt.correlation_id IS DISTINCT FROM v_job.id::text OR
       p_receipt.causation_id IS DISTINCT FROM v_job.id OR
       p_receipt.provider_reference IS DISTINCT FROM v_job.payload_ref OR
       p_receipt.retain_until IS DISTINCT FROM
           p_receipt.completed_at+interval '90 days' THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export running receipt evidence differs';
    END IF;
    IF p_receipt.effect_status='failed' THEN
        IF p_receipt.effect_error_class IS DISTINCT FROM v_expected_class OR
           p_receipt.effect_result_ref IS NOT NULL OR
           p_receipt.effect_result_sha256 IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export failure receipt differs';
        END IF;
        RETURN;
    END IF;
    v_result_state:=pg_catalog.split_part(p_receipt.effect_result_ref,'/',5);
    v_result_ref:='analytics-dashboard-export://'||v_operation||'/'||
        v_job.payload_ref||'/'||v_result_state;
    IF p_receipt.effect_error_class IS NOT NULL OR
       p_receipt.effect_result_ref IS DISTINCT FROM v_result_ref OR
       p_receipt.effect_result_sha256 IS DISTINCT FROM pg_catalog.encode(
          pg_catalog.sha256(pg_catalog.convert_to(v_result_ref,'UTF8')),'hex') OR
       (v_job.kind='analytics.dashboard.export.delete' AND
          v_result_state IS DISTINCT FROM 'deleted') OR
       (v_job.kind='analytics.dashboard.export.expire' AND
          v_result_state NOT IN ('expired','cancelled','failed','deleted')) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export success receipt differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_assert_analytics_dashboard_export_running_receipt_v1(
        platform.idempotency_receipts,name,name) FROM PUBLIC;

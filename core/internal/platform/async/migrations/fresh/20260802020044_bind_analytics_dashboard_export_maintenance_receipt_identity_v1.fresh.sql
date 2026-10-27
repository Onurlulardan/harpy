-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.platform_async_assert_analytics_dashboard_export_maintenance_receipt_v1(
    p_receipt platform.idempotency_receipts,p_current_user name,p_session_user name
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_claim platform.analytics_dashboard_export_maintenance_unknown_claims%ROWTYPE;
    v_schema name;
    v_operation text;
    v_code text;
BEGIN
    IF pg_catalog.pg_trigger_depth()<>1 OR
       p_current_user IS DISTINCT FROM 'forgecrud_owner' OR
       p_session_user IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_receipt.effect_kind NOT IN ('analytics.dashboard.export.delete',
                                     'analytics.dashboard.export.expire') THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance receipt authority is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_receipt.causation_id AND job.tenant_id=p_receipt.tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=job.tenant_id
      AND job.authority_domain='analytics' AND job.writer_origin='v2'
      AND job.kind=p_receipt.effect_kind AND job.payload_revision=1
      AND job.effect_class='external' AND job.state='dead'
      AND job.last_error_class IS NULL
      AND job.last_error_code='JOB_OUTCOME_UNKNOWN' FOR SHARE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.job_id=v_job.id AND attempt.attempt=v_job.attempts
      AND attempt.lease_generation=v_job.lease_generation
      AND attempt.outcome='dead' AND attempt.error_class IS NULL
      AND attempt.error_code='JOB_OUTCOME_UNKNOWN'
      AND attempt.effect_receipt_id IS NULL FOR SHARE;
    SELECT claim.* INTO v_claim
    FROM platform.analytics_dashboard_export_maintenance_unknown_claims AS claim
    WHERE claim.job_id=v_job.id AND claim.attempt_id=v_attempt.attempt_id
      AND claim.effect_id=v_attempt.effect_id
      AND claim.request_sha256=v_attempt.effect_request_sha256
      AND claim.claim_generation>0
      AND claim.cleanup_lease_owner_id IS NOT NULL
      AND claim.lease_until>pg_catalog.clock_timestamp() FOR SHARE;
    SELECT tenant.schema_name INTO v_schema FROM platform.tenants AS tenant
    WHERE tenant.id=v_job.tenant_id AND tenant.deleted_at IS NULL FOR SHARE;
    v_operation:=CASE v_job.kind
        WHEN 'analytics.dashboard.export.delete' THEN 'delete'
        WHEN 'analytics.dashboard.export.expire' THEN 'expire' END;
    v_code:=CASE v_job.kind
        WHEN 'analytics.dashboard.export.delete' THEN 'ANALYTICS_EXPORT_DELETED'
        WHEN 'analytics.dashboard.export.expire' THEN
            'ANALYTICS_EXPORT_EXPIRY_COMPLETE' END;
    IF v_job.id IS NULL OR v_attempt.attempt_id IS NULL OR v_claim.job_id IS NULL OR
       v_schema IS DISTINCT FROM v_claim.tenant_schema OR
       p_receipt.receipt_id IS DISTINCT FROM
          platform.platform_async_analytics_export_maint_receipt_id_v1(
              v_attempt.effect_id,v_job.kind) OR
       p_receipt.record_kind IS DISTINCT FROM 'effect' OR
       p_receipt.scope_kind IS DISTINCT FROM v_job.scope_kind OR
       p_receipt.scope_id IS DISTINCT FROM v_job.scope_id OR
       p_receipt.authority_group_id IS DISTINCT FROM v_job.authority_group_id OR
       p_receipt.authority_group_generation IS DISTINCT FROM
           v_job.authority_group_generation OR
       p_receipt.authority_domain IS DISTINCT FROM v_job.authority_domain OR
       p_receipt.cutover_generation IS DISTINCT FROM v_job.cutover_generation OR
       p_receipt.writer_origin IS DISTINCT FROM 'v2' OR
       p_receipt.request_sha256 IS DISTINCT FROM v_attempt.effect_request_sha256 OR
       p_receipt.retention_class IS DISTINCT FROM 'external_effect' OR
       p_receipt.effect_id IS DISTINCT FROM v_attempt.effect_id OR
       p_receipt.effect_revision IS DISTINCT FROM 1 OR
       p_receipt.effect_key IS DISTINCT FROM v_job.effect_idempotency_key OR
       p_receipt.provider IS DISTINCT FROM 'analytics.dashboard-export' OR
       p_receipt.operation IS DISTINCT FROM v_operation OR
       p_receipt.correlation_id IS DISTINCT FROM v_job.id::text OR
       p_receipt.effect_status IS DISTINCT FROM 'succeeded' OR
       p_receipt.effect_error_class IS NOT NULL OR
       p_receipt.effect_code IS DISTINCT FROM v_code OR
       p_receipt.provider_reference IS DISTINCT FROM v_job.payload_ref OR
       p_receipt.effect_result_ref IS NULL OR
       p_receipt.effect_result_ref IS DISTINCT FROM
          'analytics-dashboard-export://'||v_operation||'/'||v_job.payload_ref||'/'||
          pg_catalog.split_part(p_receipt.effect_result_ref,'/',5) OR
       p_receipt.effect_result_sha256 IS DISTINCT FROM pg_catalog.encode(
          pg_catalog.sha256(pg_catalog.convert_to(p_receipt.effect_result_ref,'UTF8')),
          'hex') OR p_receipt.completed_at<=v_job.completed_at OR
       p_receipt.retain_until IS DISTINCT FROM
          p_receipt.completed_at+interval '90 days' THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance receipt differs';
    END IF;
    IF (v_job.kind='analytics.dashboard.export.delete' AND
        p_receipt.effect_result_ref<> 'analytics-dashboard-export://delete/'||
            v_job.payload_ref||'/deleted') OR
       (v_job.kind='analytics.dashboard.export.expire' AND
        pg_catalog.split_part(p_receipt.effect_result_ref,'/',5) NOT IN
            ('expired','cancelled','failed','deleted')) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance receipt state differs';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_domain);
END;
$function$;

CREATE OR REPLACE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_insert_receipt_v1(
    p_job platform.jobs,p_attempt platform.job_attempts,p_effect_receipt_id uuid,
    p_effect_receipt_sha256 bytea,p_terminal_state text,p_completed_at timestamptz
)
RETURNS TABLE(provider_code text,result_reference text,result_sha256 text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE v_operation text;
BEGIN
    v_operation:=CASE p_job.kind WHEN 'analytics.dashboard.export.delete' THEN
        'delete' WHEN 'analytics.dashboard.export.expire' THEN 'expire' END;
    provider_code:=CASE p_job.kind WHEN 'analytics.dashboard.export.delete' THEN
        'ANALYTICS_EXPORT_DELETED' WHEN 'analytics.dashboard.export.expire' THEN
        'ANALYTICS_EXPORT_EXPIRY_COMPLETE' END;
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR v_operation IS NULL OR
       p_job.state IS DISTINCT FROM 'dead' OR
       p_job.last_error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       p_attempt.job_id IS DISTINCT FROM p_job.id OR
       p_attempt.outcome IS DISTINCT FROM 'dead' OR
       p_attempt.error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       p_attempt.effect_receipt_id IS NOT NULL OR
       p_effect_receipt_id IS DISTINCT FROM
          platform.platform_async_analytics_export_maint_receipt_id_v1(
              p_attempt.effect_id,p_job.kind) OR
       pg_catalog.octet_length(p_effect_receipt_sha256) IS DISTINCT FROM 32 OR
       p_effect_receipt_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       pg_catalog.isfinite(p_completed_at) IS DISTINCT FROM true OR
       p_completed_at<=p_job.completed_at OR
       p_completed_at>pg_catalog.clock_timestamp()+interval '5 seconds' OR
       (v_operation='delete' AND p_terminal_state IS DISTINCT FROM 'deleted') OR
       (v_operation='expire' AND p_terminal_state NOT IN
           ('expired','cancelled','failed','deleted')) OR
       platform.platform_async_analytics_dashboard_export_maintenance_receipt_sha256_v1(
           p_job,p_attempt,p_effect_receipt_id,p_terminal_state,p_completed_at)
           IS DISTINCT FROM p_effect_receipt_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance receipt evidence differs';
    END IF;
    result_reference:='analytics-dashboard-export://'||v_operation||'/'||
        p_job.payload_ref||'/'||p_terminal_state;
    result_sha256:=pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(result_reference,'UTF8')),'hex');
    INSERT INTO platform.idempotency_receipts(
        receipt_id,record_kind,tenant_id,scope_kind,scope_id,authority_group_id,
        authority_group_generation,authority_domain,cutover_generation,writer_origin,
        request_sha256,retention_class,effect_id,effect_kind,effect_revision,effect_key,
        provider,operation,correlation_id,causation_id,effect_status,effect_code,
        provider_reference,effect_result_ref,effect_result_sha256,completed_at,retain_until
    ) VALUES (p_effect_receipt_id,'effect',p_job.tenant_id,'tenant',p_job.tenant_id,
        p_job.authority_group_id,p_job.authority_group_generation,'analytics',
        p_job.cutover_generation,'v2',p_attempt.effect_request_sha256,'external_effect',
        p_attempt.effect_id,p_job.kind,1,p_job.effect_idempotency_key,
        'analytics.dashboard-export',v_operation,p_job.id::text,p_job.id,'succeeded',
        provider_code,p_job.payload_ref,result_reference,result_sha256,p_completed_at,
        p_completed_at+interval '90 days');
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_assert_analytics_dashboard_export_maintenance_receipt_v1(
        platform.idempotency_receipts,name,name) FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_maintenance_insert_receipt_v1(
        platform.jobs,platform.job_attempts,uuid,bytea,text,timestamptz) FROM PUBLIC;

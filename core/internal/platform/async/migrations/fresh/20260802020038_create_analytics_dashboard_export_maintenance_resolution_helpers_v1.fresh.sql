-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_insert_receipt_v1(
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
       pg_catalog.uuid_extract_version(p_effect_receipt_id) IS DISTINCT FROM 5 OR
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

CREATE FUNCTION platform.platform_async_analytics_dashboard_export_maintenance_assert_replay_v1(
    p_job platform.jobs,p_attempt platform.job_attempts,
    p_claim platform.analytics_dashboard_export_maintenance_unknown_claims,
    p_resolution platform.analytics_dashboard_export_maintenance_unknown_resolutions,
    p_effect_receipt_id uuid,p_effect_receipt_sha256 bytea,p_resolution_receipt_id uuid,
    p_resolution_sha256 bytea,p_terminal_state text,p_resolved_at timestamptz
)
RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_receipt platform.idempotency_receipts%ROWTYPE;
    v_operation text;
    v_code text;
    v_result_ref text;
BEGIN
    v_operation:=CASE p_claim.kind WHEN 'analytics.dashboard.export.delete' THEN
        'delete' WHEN 'analytics.dashboard.export.expire' THEN 'expire' END;
    v_code:=CASE p_claim.kind WHEN 'analytics.dashboard.export.delete' THEN
        'ANALYTICS_EXPORT_DELETED' ELSE 'ANALYTICS_EXPORT_EXPIRY_COMPLETE' END;
    v_result_ref:='analytics-dashboard-export://'||v_operation||'/'||
        p_claim.export_id::text||'/'||p_terminal_state;
    SELECT receipt.* INTO v_receipt FROM platform.idempotency_receipts receipt
    WHERE receipt.receipt_id=p_effect_receipt_id FOR SHARE;
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_claim.job_id IS NULL OR p_resolution.job_id IS NULL OR
       p_resolution.kind IS DISTINCT FROM p_claim.kind OR
       p_resolution.tenant_id IS DISTINCT FROM p_claim.tenant_id OR
       p_resolution.tenant_schema IS DISTINCT FROM p_claim.tenant_schema OR
       p_resolution.export_id IS DISTINCT FROM p_claim.export_id OR
       p_resolution.authority_group_id IS DISTINCT FROM p_claim.authority_group_id OR
       p_resolution.authority_group_generation IS DISTINCT FROM
           p_claim.authority_group_generation OR
       p_resolution.cutover_generation IS DISTINCT FROM p_claim.cutover_generation OR
       p_resolution.attempt_id IS DISTINCT FROM p_claim.attempt_id OR
       p_resolution.attempt_number IS DISTINCT FROM p_claim.attempt_number OR
       p_resolution.lease_generation IS DISTINCT FROM p_claim.lease_generation OR
       p_resolution.original_lease_owner_id IS DISTINCT FROM
           p_claim.original_lease_owner_id OR
       p_resolution.effect_id IS DISTINCT FROM p_claim.effect_id OR
       p_resolution.request_sha256 IS DISTINCT FROM p_claim.request_sha256 OR
       p_resolution.original_unknown_at IS DISTINCT FROM p_claim.unknown_at OR
       p_resolution.original_attempt_finished_at IS DISTINCT FROM p_claim.unknown_at OR
       p_resolution.original_attempt_outcome IS DISTINCT FROM 'dead' OR
       p_resolution.original_error_class IS NOT NULL OR
       p_resolution.original_error_code IS DISTINCT FROM 'JOB_OUTCOME_UNKNOWN' OR
       p_resolution.original_error_sha256 IS NOT NULL OR
       p_resolution.cleanup_generation IS DISTINCT FROM p_claim.claim_generation OR
       p_resolution.cleanup_lease_owner_id IS DISTINCT FROM
           p_claim.cleanup_lease_owner_id OR
       p_resolution.resolution_receipt_id IS DISTINCT FROM p_resolution_receipt_id OR
       p_resolution.effect_receipt_id IS DISTINCT FROM p_effect_receipt_id OR
       p_resolution.effect_receipt_sha256 IS DISTINCT FROM
           pg_catalog.encode(p_effect_receipt_sha256,'hex') OR
       p_resolution.resolution_sha256 IS DISTINCT FROM
           pg_catalog.encode(p_resolution_sha256,'hex') OR
       p_resolution.terminal_state IS DISTINCT FROM p_terminal_state OR
       p_resolution.provider_code IS DISTINCT FROM v_code OR
       p_resolution.result_reference IS DISTINCT FROM v_result_ref OR
       p_resolution.resolved_at IS DISTINCT FROM p_resolved_at OR
       p_job.id IS DISTINCT FROM p_claim.job_id OR p_job.state IS DISTINCT FROM 'succeeded' OR
       p_job.completed_at IS DISTINCT FROM p_resolved_at OR
       p_job.updated_at IS DISTINCT FROM p_resolved_at OR
       p_job.retain_until IS DISTINCT FROM p_resolved_at+interval '90 days' OR
       p_attempt.attempt_id IS DISTINCT FROM p_claim.attempt_id OR
       p_attempt.outcome IS DISTINCT FROM 'succeeded' OR
       p_attempt.effect_receipt_id IS DISTINCT FROM p_effect_receipt_id OR
       p_attempt.finished_at IS DISTINCT FROM p_resolved_at OR
       v_receipt.receipt_id IS NULL OR v_receipt.effect_id IS DISTINCT FROM p_claim.effect_id OR
       v_receipt.effect_kind IS DISTINCT FROM p_claim.kind OR
       v_receipt.request_sha256 IS DISTINCT FROM p_claim.request_sha256 OR
       v_receipt.effect_status IS DISTINCT FROM 'succeeded' OR
       v_receipt.effect_code IS DISTINCT FROM v_code OR
       v_receipt.effect_result_ref IS DISTINCT FROM v_result_ref OR
       v_receipt.completed_at IS DISTINCT FROM p_resolved_at OR
       platform.platform_async_analytics_dashboard_export_maintenance_receipt_sha256_v1(
           p_job,p_attempt,p_effect_receipt_id,p_terminal_state,p_resolved_at)
           IS DISTINCT FROM p_effect_receipt_sha256 THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export maintenance resolution replay differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_maintenance_insert_receipt_v1(
        platform.jobs,platform.job_attempts,uuid,bytea,text,timestamptz) FROM PUBLIC;
REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_maintenance_assert_replay_v1(
        platform.jobs,platform.job_attempts,
        platform.analytics_dashboard_export_maintenance_unknown_claims,
        platform.analytics_dashboard_export_maintenance_unknown_resolutions,
        uuid,bytea,uuid,bytea,text,timestamptz) FROM PUBLIC;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_effect_receipt_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_receipt_id uuid,p_status text,p_error_class text,p_code text,
    p_provider_reference text,p_result_sha256 bytea,p_completed_at timestamptz
)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_receipt_id uuid;
    v_result_ref text;
    v_result_digest text;
    v_exact boolean;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       (p_receipt_id IS NOT NULL AND
        pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7) OR
       p_status NOT IN ('succeeded','failed') OR
       p_code !~ '^[A-Z][A-Z0-9_]{1,63}$' OR
       (p_provider_reference IS NOT NULL AND
        pg_catalog.octet_length(p_provider_reference) NOT BETWEEN 1 AND 512) OR
       NOT pg_catalog.isfinite(p_completed_at) OR
       (p_status='succeeded' AND (p_receipt_id IS NULL OR
        p_error_class IS NOT NULL OR
        pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
        p_result_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex'))) OR
       (p_status='failed' AND (
        p_error_class NOT IN ('transient','permanent','conflict','cancelled') OR
        p_result_sha256 IS NOT NULL)) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export effect receipt input is invalid';
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
      AND job.effect_idempotency_key=job.dedupe_key;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id;
    IF v_job.id IS NULL OR v_attempt.attempt_id IS NULL OR
       v_attempt.effect_started_at IS NULL OR v_attempt.effect_id IS NULL OR
       v_attempt.effect_request_sha256 IS NULL OR
       p_completed_at<v_attempt.effect_started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export effect evidence is incomplete';
    END IF;
    v_receipt_id:=COALESCE(p_receipt_id,pg_catalog.uuidv7());
    v_result_ref:=CASE WHEN p_status='succeeded' THEN
        'analytics-dashboard-export/'||p_export_id::text||'/artifact' END;
    v_result_digest:=CASE WHEN p_status='succeeded' THEN
        pg_catalog.encode(p_result_sha256,'hex') END;
    INSERT INTO platform.idempotency_receipts(
        receipt_id,record_kind,tenant_id,scope_kind,scope_id,
        authority_group_id,authority_group_generation,authority_domain,
        cutover_generation,writer_origin,request_sha256,retention_class,
        effect_id,effect_kind,effect_revision,effect_key,provider,operation,
        correlation_id,causation_id,effect_status,effect_error_class,effect_code,
        provider_reference,effect_result_ref,effect_result_sha256,
        completed_at,retain_until
    ) VALUES (
        v_receipt_id,'effect',p_tenant_id,'tenant',p_tenant_id,
        v_job.authority_group_id,v_job.authority_group_generation,
        v_job.authority_domain,v_job.cutover_generation,'v2',
        v_attempt.effect_request_sha256,'external_effect',v_attempt.effect_id,
        'analytics.dashboard.export.render',1,v_job.effect_idempotency_key,
        'object-storage','put-object',v_attempt.effect_id::text,p_job_id,
        p_status,p_error_class,p_code,p_provider_reference,v_result_ref,
        v_result_digest,p_completed_at,p_completed_at+interval '91 days'
    ) ON CONFLICT DO NOTHING;
    SELECT receipt.receipt_id=v_receipt_id INTO v_exact
    FROM platform.idempotency_receipts AS receipt
    WHERE receipt.receipt_id=v_receipt_id AND receipt.record_kind='effect'
      AND receipt.scope_kind='tenant' AND receipt.scope_id=p_tenant_id
      AND receipt.authority_domain='analytics'
      AND receipt.effect_id=v_attempt.effect_id
      AND receipt.effect_kind='analytics.dashboard.export.render'
      AND receipt.effect_revision=1 AND receipt.effect_key=v_job.dedupe_key
      AND receipt.provider='object-storage' AND receipt.operation='put-object'
      AND receipt.request_sha256=v_attempt.effect_request_sha256
      AND receipt.effect_status=p_status
      AND receipt.effect_error_class IS NOT DISTINCT FROM p_error_class
      AND receipt.effect_code=p_code
      AND receipt.provider_reference IS NOT DISTINCT FROM p_provider_reference
      AND receipt.effect_result_ref IS NOT DISTINCT FROM v_result_ref
      AND receipt.effect_result_sha256 IS NOT DISTINCT FROM v_result_digest
      AND receipt.completed_at=p_completed_at;
    IF NOT COALESCE(v_exact,false) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export effect receipt replay differs';
    END IF;
    RETURN v_receipt_id;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_effect_receipt_v1(
        uuid,uuid,uuid,uuid,uuid,text,text,text,text,bytea,timestamptz)
FROM PUBLIC;

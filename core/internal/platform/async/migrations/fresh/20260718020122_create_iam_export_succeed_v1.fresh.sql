-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_export_succeed_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_lease_owner_id uuid,p_lease_generation bigint,p_resource_kind text,
    p_effect_id uuid,p_receipt_id uuid,p_request_sha256 bytea,
    p_result_sha256 bytea,p_provider_reference text,p_completed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_cancelled boolean;
    v_request text;
    v_result text;
    v_receipt_exact boolean;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_provider_reference) NOT BETWEEN 1 AND 512 OR
       NOT pg_catalog.isfinite(p_completed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM export success evidence is invalid';
    END IF;
    v_request := pg_catalog.encode(p_request_sha256,'hex');
    v_result := pg_catalog.encode(p_result_sha256,'hex');
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_domain='iam' AND job.writer_origin='v2'
      AND job.authority_group_id='iam_system_resources'
      AND job.kind='iam.system-resource.record-export'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.payload_ref='iam-system-record/'||p_resource_kind||'/'||p_export_id::text
      AND job.effect_idempotency_key='iam-system-export:'||p_export_id::text
      AND job.dedupe_key='iam-system-transfer:'||
          'iam.system-resource.record-export:'||p_export_id::text
      AND job.max_attempts=3
    FOR UPDATE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
      AND attempt.attempt=p_lease_generation
      AND attempt.lease_generation=p_lease_generation
      AND attempt.lease_owner_id=p_lease_owner_id FOR UPDATE;
    IF v_job.state='succeeded' AND v_attempt.outcome='succeeded' THEN
        SELECT receipt.receipt_id=p_receipt_id INTO v_receipt_exact
        FROM platform.idempotency_receipts AS receipt
        WHERE receipt.receipt_id=p_receipt_id AND receipt.record_kind='effect'
          AND receipt.scope_kind='tenant' AND receipt.scope_id=p_tenant_id
          AND receipt.effect_id=p_effect_id AND receipt.request_sha256=v_request
          AND receipt.effect_status='succeeded'
          AND receipt.effect_code='OBJECT_UPLOAD_SUCCEEDED'
          AND receipt.provider_reference=p_provider_reference
          AND receipt.effect_result_sha256=v_result
          AND receipt.completed_at=p_completed_at;
        IF COALESCE(v_receipt_exact,false) AND
           v_attempt.effect_id=p_effect_id AND
           v_attempt.effect_request_sha256=v_request AND
           v_attempt.effect_receipt_id=p_receipt_id AND
           v_attempt.finished_at=p_completed_at AND
           v_job.completed_at=p_completed_at THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export success replay differs';
    END IF;
    SELECT platform.platform_async_iam_export_current_lease_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind
    ) INTO v_cancelled;
    IF v_cancelled THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='IAM export cancellation was requested';
    END IF;
    IF v_attempt.effect_id IS DISTINCT FROM p_effect_id OR
       v_attempt.effect_request_sha256 IS DISTINCT FROM v_request OR
       v_attempt.effect_started_at IS NULL OR
       p_completed_at<v_attempt.effect_started_at OR
       p_completed_at<=v_job.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export success lease is stale';
    END IF;
    INSERT INTO platform.idempotency_receipts(
        receipt_id,record_kind,tenant_id,scope_kind,scope_id,
        authority_group_id,authority_group_generation,authority_domain,
        cutover_generation,writer_origin,request_sha256,retention_class,
        effect_id,effect_kind,effect_revision,effect_key,provider,operation,
        correlation_id,causation_id,effect_status,effect_code,provider_reference,
        effect_result_ref,effect_result_sha256,completed_at,retain_until
    ) VALUES(
        p_receipt_id,'effect',p_tenant_id,'tenant',p_tenant_id,
        v_job.authority_group_id,v_job.authority_group_generation,
        v_job.authority_domain,v_job.cutover_generation,'v2',v_request,
        'external_effect',p_effect_id,'iam.system-export.artifact',1,
        v_job.effect_idempotency_key,'object-storage','put-object',
        p_effect_id::text,p_job_id,'succeeded','OBJECT_UPLOAD_SUCCEEDED',
        p_provider_reference,'iam-system-export/'||p_export_id::text||'/artifact',
        v_result,p_completed_at,p_completed_at+interval '91 days'
    );
    UPDATE platform.job_attempts SET finished_at=p_completed_at,
        outcome='succeeded',effect_receipt_id=p_receipt_id
    WHERE attempt_id=p_attempt_id;
    UPDATE platform.jobs SET state='succeeded',lease_owner_id=NULL,
        lease_heartbeat_at=NULL,lease_until=NULL,last_error_class=NULL,
        last_error_code=NULL,last_error_sha256=NULL,
        retain_until=p_completed_at+interval '91 days',updated_at=p_completed_at,
        completed_at=p_completed_at WHERE id=p_job_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_iam_export_succeed_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,bytea,text,timestamptz
) FROM PUBLIC;

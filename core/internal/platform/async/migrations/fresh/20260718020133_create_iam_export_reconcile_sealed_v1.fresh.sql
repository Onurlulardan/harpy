-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_export_reconcile_sealed_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_lease_owner_id uuid,p_lease_generation bigint,p_resource_kind text,
    p_effect_id uuid,p_receipt_id uuid,p_request_sha256 bytea,
    p_result_sha256 bytea,p_provider_reference text,p_observed_at timestamptz,
    p_original_job_id uuid,p_original_effect_id uuid,p_original_receipt_id uuid,
    p_original_request_sha256 bytea,p_artifact_result_sha256 bytea,
    p_artifact_provider_reference text,p_artifact_sealed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_original platform.jobs%ROWTYPE;
    v_cancelled boolean;
    v_request text;
    v_result text;
    v_original_request text;
    v_artifact_result text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_original_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_original_receipt_id) IS DISTINCT FROM 7 OR
       p_effect_id=p_receipt_id OR p_original_effect_id=p_original_receipt_id OR
       p_receipt_id=p_original_receipt_id OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_result_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_original_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_artifact_result_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_provider_reference) NOT BETWEEN 1 AND 512 OR
       pg_catalog.octet_length(p_artifact_provider_reference) NOT BETWEEN 1 AND 512 OR
       NOT pg_catalog.isfinite(p_observed_at) OR
       NOT pg_catalog.isfinite(p_artifact_sealed_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM export sealed reconciliation is invalid';
    END IF;
    v_request := pg_catalog.encode(p_request_sha256,'hex');
    v_result := pg_catalog.encode(p_result_sha256,'hex');
    v_original_request := pg_catalog.encode(p_original_request_sha256,'hex');
    v_artifact_result := pg_catalog.encode(p_artifact_result_sha256,'hex');
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_group_id='iam_system_resources'
      AND job.authority_domain='iam' AND job.writer_origin='v2'
      AND job.kind='iam.system-resource.export-reconcile'
      AND job.payload_revision=1
      AND job.payload_ref='iam-system-record/'||p_resource_kind||'/'||p_export_id::text
      AND job.dedupe_key='iam-system-transfer:'||
          'iam.system-resource.export-reconcile:'||p_export_id::text
      AND job.effect_class='external'
      AND job.effect_idempotency_key='iam-system-export-reconcile:'||p_export_id::text
      AND job.max_attempts=8 FOR UPDATE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
      AND attempt.attempt=p_lease_generation
      AND attempt.lease_generation=p_lease_generation
      AND attempt.lease_owner_id=p_lease_owner_id FOR UPDATE;
    SELECT original.* INTO v_original FROM platform.jobs AS original
    WHERE original.id=p_original_job_id AND original.tenant_id=p_tenant_id
      AND original.kind='iam.system-resource.record-export'
      AND original.payload_ref='iam-system-record/'||p_resource_kind||'/'||p_export_id::text
      AND original.effect_idempotency_key='iam-system-export:'||p_export_id::text
      AND original.state='dead' AND original.last_error_class IS NULL
      AND original.last_error_code='JOB_OUTCOME_UNKNOWN' FOR SHARE;
    IF v_job.state='succeeded' AND v_attempt.outcome='succeeded' THEN
        IF v_attempt.effect_id=p_effect_id AND
           v_attempt.effect_request_sha256=v_request AND
           v_attempt.effect_receipt_id=p_receipt_id AND
           v_attempt.finished_at=p_observed_at AND v_job.completed_at=p_observed_at AND
           EXISTS(SELECT 1 FROM platform.idempotency_receipts AS receipt
             WHERE receipt.receipt_id=p_receipt_id AND receipt.effect_id=p_effect_id
               AND receipt.request_sha256=v_request AND receipt.effect_status='succeeded'
               AND receipt.effect_code='OBJECT_ARTIFACT_SEALED'
               AND receipt.provider_reference=p_provider_reference
               AND receipt.effect_result_sha256=v_result
               AND receipt.completed_at=p_observed_at) AND
           EXISTS(SELECT 1 FROM platform.idempotency_receipts AS receipt
             WHERE receipt.receipt_id=p_original_receipt_id
               AND receipt.effect_id=p_original_effect_id
               AND receipt.request_sha256=v_original_request
               AND receipt.effect_status='succeeded'
               AND receipt.effect_code='OBJECT_UPLOAD_RECONCILED'
               AND receipt.provider_reference=p_artifact_provider_reference
               AND receipt.effect_result_sha256=v_artifact_result
               AND receipt.completed_at=p_artifact_sealed_at) THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export sealed reconciliation replay differs';
    END IF;
    IF v_original.id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export original unknown job is unavailable';
    END IF;
    SELECT platform.platform_async_iam_export_reconcile_lease_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind
    ) INTO v_cancelled;
    IF v_cancelled THEN
        RAISE EXCEPTION USING ERRCODE='P6046',
            MESSAGE='IAM export reconciliation cancellation was requested';
    END IF;
    IF v_attempt.effect_id IS DISTINCT FROM p_effect_id OR
       v_attempt.effect_request_sha256 IS DISTINCT FROM v_request OR
       v_attempt.effect_started_at IS NULL OR p_observed_at<v_attempt.effect_started_at OR
       p_observed_at<v_job.updated_at OR
       p_artifact_sealed_at<v_original.effect_started_at-interval '1 second' OR
       EXISTS(SELECT 1 FROM platform.idempotency_receipts AS receipt
              WHERE receipt.receipt_id IN (p_receipt_id,p_original_receipt_id)) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export sealed reconciliation evidence conflicts';
    END IF;
    INSERT INTO platform.idempotency_receipts(
        receipt_id,record_kind,tenant_id,scope_kind,scope_id,authority_group_id,
        authority_group_generation,authority_domain,cutover_generation,writer_origin,
        request_sha256,retention_class,effect_id,effect_kind,effect_revision,
        effect_key,provider,operation,correlation_id,causation_id,effect_status,
        effect_code,provider_reference,effect_result_ref,effect_result_sha256,
        completed_at,retain_until
    ) VALUES
    (p_original_receipt_id,'effect',p_tenant_id,'tenant',p_tenant_id,
     v_original.authority_group_id,v_original.authority_group_generation,
     v_original.authority_domain,v_original.cutover_generation,'v2',v_original_request,
     'external_effect',p_original_effect_id,'iam.system-export.artifact',1,
     v_original.effect_idempotency_key,'object-storage','head-object',
     p_original_effect_id::text,p_original_job_id,'succeeded','OBJECT_UPLOAD_RECONCILED',
     p_artifact_provider_reference,'iam-system-export/'||p_export_id::text||'/artifact',
     v_artifact_result,p_artifact_sealed_at,p_observed_at+interval '91 days'),
    (p_receipt_id,'effect',p_tenant_id,'tenant',p_tenant_id,
     v_job.authority_group_id,v_job.authority_group_generation,v_job.authority_domain,
     v_job.cutover_generation,'v2',v_request,'external_effect',p_effect_id,
     'iam.system-resource.export-reconcile',1,v_job.effect_idempotency_key,
     'object-storage','head-object',p_effect_id::text,p_job_id,'succeeded',
     'OBJECT_ARTIFACT_SEALED',p_provider_reference,
     'iam-system-export/'||p_export_id::text||'/reconciliation',v_result,
     p_observed_at,p_observed_at+interval '91 days');
    UPDATE platform.job_attempts SET finished_at=p_observed_at,
        outcome='succeeded',effect_receipt_id=p_receipt_id WHERE attempt_id=p_attempt_id;
    UPDATE platform.jobs SET state='succeeded',lease_owner_id=NULL,
        lease_heartbeat_at=NULL,lease_until=NULL,last_error_class=NULL,
        last_error_code=NULL,last_error_sha256=NULL,
        retain_until=p_observed_at+interval '91 days',updated_at=p_observed_at,
        completed_at=p_observed_at WHERE id=p_job_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_iam_export_reconcile_sealed_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,bytea,text,timestamptz,
    uuid,uuid,uuid,bytea,bytea,text,timestamptz
) FROM PUBLIC;

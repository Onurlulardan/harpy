-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_export_reconcile_retry_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_lease_owner_id uuid,p_lease_generation bigint,p_resource_kind text,
    p_effect_id uuid,p_receipt_id uuid,p_request_sha256 bytea,
    p_failure_sha256 bytea,p_observed_at timestamptz,p_retry_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_cancelled boolean;
    v_request text;
    v_failure text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_lease_generation NOT BETWEEN 1 AND 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_effect_id=p_receipt_id OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       pg_catalog.octet_length(p_failure_sha256) IS DISTINCT FROM 32 OR
       NOT pg_catalog.isfinite(p_observed_at) OR NOT pg_catalog.isfinite(p_retry_at) OR
       p_retry_at<p_observed_at OR p_retry_at>p_observed_at+interval '15 minutes' THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM export reconciliation retry is invalid';
    END IF;
    v_request := pg_catalog.encode(p_request_sha256,'hex');
    v_failure := pg_catalog.encode(p_failure_sha256,'hex');
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_group_id='iam_system_resources'
      AND job.authority_domain='iam' AND job.writer_origin='v2'
      AND job.kind='iam.system-resource.export-reconcile'
      AND job.payload_ref='iam-system-record/'||p_resource_kind||'/'||p_export_id::text
      AND job.effect_idempotency_key='iam-system-export-reconcile:'||p_export_id::text
      AND job.max_attempts=8 FOR UPDATE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.attempt=p_lease_generation
      AND attempt.lease_generation=p_lease_generation
      AND attempt.lease_owner_id=p_lease_owner_id FOR UPDATE;
    IF v_job.state='queued' AND v_attempt.outcome='requeued' THEN
        IF v_job.run_at=p_retry_at AND v_job.updated_at=p_observed_at AND
           v_job.last_error_class='transient' AND
           v_job.last_error_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN' AND
           v_job.last_error_sha256=v_failure AND
           v_attempt.effect_id=p_effect_id AND
           v_attempt.effect_request_sha256=v_request AND
           v_attempt.effect_receipt_id=p_receipt_id AND
           v_attempt.error_class='transient' AND
           v_attempt.error_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN' AND
           v_attempt.redacted_error_sha256=v_failure AND
           v_attempt.finished_at=p_observed_at AND EXISTS(
               SELECT 1 FROM platform.idempotency_receipts AS receipt
               WHERE receipt.receipt_id=p_receipt_id AND receipt.effect_id=p_effect_id
                 AND receipt.request_sha256=v_request AND receipt.effect_status='failed'
                 AND receipt.effect_error_class='transient'
                 AND receipt.effect_code=
                     'IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN'
                 AND receipt.completed_at=p_observed_at
           ) THEN RETURN true; END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export reconciliation retry replay differs';
    END IF;
    SELECT platform.platform_async_iam_export_reconcile_lease_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind
    ) INTO v_cancelled;
    IF v_cancelled OR v_attempt.effect_id IS DISTINCT FROM p_effect_id OR
       v_attempt.effect_request_sha256 IS DISTINCT FROM v_request OR
       v_attempt.effect_started_at IS NULL OR p_observed_at<v_attempt.effect_started_at OR
       p_observed_at<v_job.updated_at OR EXISTS(
           SELECT 1 FROM platform.idempotency_receipts AS receipt
           WHERE receipt.receipt_id=p_receipt_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export reconciliation retry evidence conflicts';
    END IF;
    INSERT INTO platform.idempotency_receipts(
        receipt_id,record_kind,tenant_id,scope_kind,scope_id,authority_group_id,
        authority_group_generation,authority_domain,cutover_generation,writer_origin,
        request_sha256,retention_class,effect_id,effect_kind,effect_revision,
        effect_key,provider,operation,correlation_id,causation_id,effect_status,
        effect_error_class,effect_code,completed_at,retain_until
    ) VALUES(
        p_receipt_id,'effect',p_tenant_id,'tenant',p_tenant_id,
        v_job.authority_group_id,v_job.authority_group_generation,
        v_job.authority_domain,v_job.cutover_generation,'v2',v_request,
        'external_effect',p_effect_id,'iam.system-resource.export-reconcile',1,
        v_job.effect_idempotency_key,'object-storage','head-object',p_effect_id::text,
        p_job_id,'failed','transient',
        'IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN',p_observed_at,
        p_observed_at+interval '91 days'
    );
    UPDATE platform.job_attempts SET finished_at=p_observed_at,outcome='requeued',
        error_class='transient',
        error_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN',
        redacted_error_sha256=v_failure,effect_receipt_id=p_receipt_id
    WHERE attempt_id=p_attempt_id;
    UPDATE platform.jobs SET state='queued',run_at=p_retry_at,lease_owner_id=NULL,
        lease_heartbeat_at=NULL,lease_until=NULL,last_error_class='transient',
        last_error_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN',
        last_error_sha256=v_failure,updated_at=p_observed_at WHERE id=p_job_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_iam_export_reconcile_retry_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,bytea,timestamptz,
    timestamptz
) FROM PUBLIC;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_export_finish_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_lease_owner_id uuid,p_lease_generation bigint,p_resource_kind text,
    p_failure_class text,p_failure_code text,p_failure_sha256 bytea,
    p_finished_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_cancelled boolean;
    v_digest text;
    v_job_state text;
    v_outcome text;
    v_receipt_id uuid;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_failure_class NOT IN ('transient','permanent','conflict','cancelled') OR
       p_failure_code !~ '^[A-Z][A-Z0-9_]{1,63}$' OR
       pg_catalog.octet_length(p_failure_sha256) IS DISTINCT FROM 32 OR
       NOT pg_catalog.isfinite(p_finished_at) OR
       (p_failure_class='transient' AND p_lease_generation<>3) OR
       (p_failure_class='cancelled' AND
        p_failure_code<>'IAM_SYSTEM_EXPORT_CANCELLED') THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM export terminal failure is invalid';
    END IF;
    v_digest := pg_catalog.encode(p_failure_sha256,'hex');
    v_job_state := CASE WHEN p_failure_class='cancelled'
        THEN 'cancelled' ELSE 'dead' END;
    v_outcome := v_job_state;
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
    IF v_job.state=v_job_state AND v_attempt.outcome=v_outcome THEN
        IF v_attempt.error_class=p_failure_class AND
           v_attempt.error_code=p_failure_code AND
           v_attempt.redacted_error_sha256=v_digest AND
           v_attempt.finished_at=p_finished_at AND
           v_job.last_error_class=p_failure_class AND
           v_job.last_error_code=p_failure_code AND
           v_job.last_error_sha256=v_digest AND
           v_job.completed_at=p_finished_at THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export terminal replay differs';
    END IF;
    SELECT platform.platform_async_iam_export_current_lease_v1(
        p_tenant_id,p_export_id,p_job_id,p_attempt_id,p_lease_owner_id,
        p_lease_generation,p_resource_kind
    ) INTO v_cancelled;
    IF v_cancelled IS DISTINCT FROM (p_failure_class='cancelled') OR
       p_finished_at<=v_job.updated_at OR p_finished_at<v_attempt.started_at THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export terminal lease is stale';
    END IF;
    IF v_attempt.effect_started_at IS NOT NULL THEN
        v_receipt_id := pg_catalog.uuidv7();
        INSERT INTO platform.idempotency_receipts(
            receipt_id,record_kind,tenant_id,scope_kind,scope_id,
            authority_group_id,authority_group_generation,authority_domain,
            cutover_generation,writer_origin,request_sha256,retention_class,
            effect_id,effect_kind,effect_revision,effect_key,provider,operation,
            correlation_id,causation_id,effect_status,effect_error_class,
            effect_code,completed_at,retain_until
        ) VALUES(
            v_receipt_id,'effect',p_tenant_id,'tenant',p_tenant_id,
            v_job.authority_group_id,v_job.authority_group_generation,
            v_job.authority_domain,v_job.cutover_generation,'v2',
            v_attempt.effect_request_sha256,'external_effect',v_attempt.effect_id,
            'iam.system-export.artifact',1,v_job.effect_idempotency_key,
            'object-storage','put-object',v_attempt.effect_id::text,p_job_id,
            'failed',p_failure_class,p_failure_code,p_finished_at,
            p_finished_at+interval '91 days'
        );
    END IF;
    UPDATE platform.job_attempts SET finished_at=p_finished_at,
        outcome=v_outcome,error_class=p_failure_class,error_code=p_failure_code,
        redacted_error_sha256=v_digest,effect_receipt_id=v_receipt_id
    WHERE attempt_id=p_attempt_id;
    UPDATE platform.jobs SET state=v_job_state,lease_owner_id=NULL,
        lease_heartbeat_at=NULL,lease_until=NULL,last_error_class=p_failure_class,
        last_error_code=p_failure_code,last_error_sha256=v_digest,
        retain_until=p_finished_at+CASE WHEN v_job_state='dead'
            THEN interval '366 days' ELSE interval '91 days' END,
        updated_at=p_finished_at,completed_at=p_finished_at WHERE id=p_job_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_iam_export_finish_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,text,text,bytea,timestamptz
) FROM PUBLIC;

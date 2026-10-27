-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_iam_export_reconcile_expired_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_lease_owner_id uuid,p_lease_generation bigint,p_resource_kind text,
    p_effect_id uuid,p_receipt_id uuid,p_request_sha256 bytea,
    p_effect_started_at timestamptz,p_recovered_at timestamptz,
    p_retry_at timestamptz,p_code text
)
RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_request text;
    v_failure text;
    v_class text;
    v_state text;
    v_outcome text;
    v_current boolean;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       p_lease_generation NOT BETWEEN 1 AND 8 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_receipt_id) IS DISTINCT FROM 7 OR
       p_effect_id=p_receipt_id OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       NOT pg_catalog.isfinite(p_effect_started_at) OR
       NOT pg_catalog.isfinite(p_recovered_at) OR
       p_recovered_at<p_effect_started_at OR
       (p_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN' AND (
           p_lease_generation=8 OR
           NOT pg_catalog.isfinite(p_retry_at) OR p_retry_at<p_recovered_at OR
           p_retry_at>p_recovered_at+LEAST(
               interval '5 seconds' * pg_catalog.power(
                   2::double precision,
                   (p_lease_generation-1)::double precision),
               interval '15 minutes'))) OR
       (p_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED' AND
           p_retry_at IS NOT NULL) OR
       p_code NOT IN ('IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN',
           'IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED') THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='IAM export expired reconciliation is invalid';
    END IF;
    v_request := pg_catalog.encode(p_request_sha256,'hex');
    v_failure := pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to(
            'iam.system-export.reconciliation-failure.v1','UTF8') ||
        pg_catalog.decode('00','hex') || pg_catalog.uuid_send(p_job_id) ||
        pg_catalog.uuid_send(p_effect_id) ||
        pg_catalog.int8send(pg_catalog.octet_length(
            pg_catalog.convert_to(p_code,'UTF8'))::bigint) ||
        pg_catalog.convert_to(p_code,'UTF8') ||
        pg_catalog.int8send(p_lease_generation)),'hex');
    v_class := CASE WHEN p_code=
        'IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN'
        THEN 'transient' ELSE 'permanent' END;
    v_state := CASE WHEN v_class='transient' THEN 'queued' ELSE 'dead' END;
    v_outcome := CASE WHEN v_class='transient' THEN 'requeued' ELSE 'dead' END;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.authority_group_id='iam_system_resources'
      AND job.authority_domain='iam' AND job.writer_origin='v2'
      AND job.kind='iam.system-resource.export-reconcile'
      AND job.payload_revision=1 AND job.effect_class='external'
      AND job.payload_ref='iam-system-record/'||p_resource_kind||'/'||p_export_id::text
      AND job.dedupe_key='iam-system-transfer:'||
          'iam.system-resource.export-reconcile:'||p_export_id::text
      AND job.effect_idempotency_key='iam-system-export-reconcile:'||p_export_id::text
      AND job.max_attempts=8 FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export expired reconciliation job is stale';
    END IF;
    SELECT EXISTS(
        SELECT 1 FROM platform.cutover_groups AS group_row
        JOIN platform.domain_cutovers AS domain_row
          ON domain_row.scope_kind=group_row.scope_kind
         AND domain_row.scope_id=group_row.scope_id
        WHERE group_row.scope_kind=v_job.scope_kind
          AND group_row.scope_id=v_job.scope_id
          AND group_row.group_id=v_job.authority_group_id
          AND group_row.generation=v_job.authority_group_generation
          AND group_row.state='v2' AND group_row.writer='v2'
          AND domain_row.domain=v_job.authority_domain
          AND domain_row.generation=v_job.cutover_generation
          AND domain_row.state='v2' AND domain_row.writer='v2'
    ) INTO v_current;
    IF (p_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_UNKNOWN' AND
        NOT v_current) OR
       (p_code='IAM_SYSTEM_EXPORT_ARTIFACT_RECONCILIATION_PARKED' AND
        p_lease_generation<8 AND v_current) THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export expired reconciliation disposition conflicts';
    END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
      AND attempt.attempt=p_lease_generation
      AND attempt.lease_generation=p_lease_generation
      AND attempt.lease_owner_id=p_lease_owner_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export expired reconciliation attempt is stale';
    END IF;
    IF v_job.state=v_state AND v_attempt.outcome=v_outcome THEN
        IF v_job.last_error_class=v_class AND v_job.last_error_code=p_code AND
           v_job.last_error_sha256=v_failure AND v_job.updated_at=p_recovered_at AND
           ((v_state='queued' AND v_job.run_at=p_retry_at AND
             v_job.completed_at IS NULL) OR
            (v_state='dead' AND v_job.completed_at=p_recovered_at)) AND
           v_attempt.effect_id=p_effect_id AND
           v_attempt.effect_request_sha256=v_request AND
           v_attempt.effect_started_at=p_effect_started_at AND
           v_attempt.effect_receipt_id=p_receipt_id AND
           v_attempt.error_class=v_class AND v_attempt.error_code=p_code AND
           v_attempt.redacted_error_sha256=v_failure AND
           v_attempt.finished_at=p_recovered_at AND EXISTS(
               SELECT 1 FROM platform.idempotency_receipts AS receipt
               WHERE receipt.receipt_id=p_receipt_id AND receipt.effect_id=p_effect_id
                 AND receipt.request_sha256=v_request AND receipt.effect_status='failed'
                 AND receipt.effect_error_class=v_class AND receipt.effect_code=p_code
                 AND receipt.completed_at=p_recovered_at
           ) THEN RETURN true; END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='IAM export expired reconciliation replay differs';
    END IF;
    IF v_job.state<>'running' OR v_job.attempts<>p_lease_generation OR
       v_job.lease_generation<>p_lease_generation OR
       v_job.lease_owner_id<>p_lease_owner_id OR
       v_job.lease_until>p_recovered_at OR v_job.updated_at>=p_recovered_at OR
       v_attempt.finished_at IS NOT NULL OR
       v_attempt.effect_id IS DISTINCT FROM p_effect_id OR
       v_attempt.effect_request_sha256 IS DISTINCT FROM v_request OR
       v_attempt.effect_started_at IS DISTINCT FROM p_effect_started_at OR
       v_attempt.effect_receipt_id IS NOT NULL OR EXISTS(
           SELECT 1 FROM platform.idempotency_receipts AS receipt
           WHERE receipt.receipt_id=p_receipt_id
       ) THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='IAM export expired reconciliation lease is stale';
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
        p_job_id,'failed',v_class,p_code,p_recovered_at,
        p_recovered_at+CASE WHEN v_state='dead'
            THEN interval '366 days' ELSE interval '91 days' END
    );
    UPDATE platform.job_attempts SET finished_at=p_recovered_at,outcome=v_outcome,
        error_class=v_class,error_code=p_code,redacted_error_sha256=v_failure,
        effect_receipt_id=p_receipt_id WHERE attempt_id=p_attempt_id;
    UPDATE platform.jobs SET state=v_state,run_at=COALESCE(p_retry_at,run_at),
        lease_owner_id=NULL,lease_heartbeat_at=NULL,lease_until=NULL,
        last_error_class=v_class,last_error_code=p_code,last_error_sha256=v_failure,
        retain_until=CASE WHEN v_state='dead'
            THEN p_recovered_at+interval '366 days' END,
        updated_at=p_recovered_at,
        completed_at=CASE WHEN v_state='dead' THEN p_recovered_at END
    WHERE id=p_job_id;
    RETURN false;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_iam_export_reconcile_expired_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,uuid,uuid,bytea,timestamptz,
    timestamptz,timestamptz,text
) FROM PUBLIC;

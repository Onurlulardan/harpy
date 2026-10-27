-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_finish_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_attempt_generation bigint,p_lease_owner_id uuid,p_failure_class text,
    p_failure_code text,p_failure_sha256 bytea,p_retryable boolean,
    p_provider_reference text,p_finished_at timestamptz
)
RETURNS TABLE(job_state text,effect_receipt_id uuid)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_receipt_id uuid;
    v_outcome text;
    v_job_state text;
    v_retry_at timestamptz;
    v_digest text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export finish authority is invalid';
    END IF;
    IF p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_lease_owner_id) IS DISTINCT FROM 7 OR
       p_failure_class NOT IN ('transient','permanent','conflict','cancelled') OR
       p_failure_code !~ '^[A-Z][A-Z0-9_]{1,63}$' OR
       pg_catalog.octet_length(p_failure_sha256) IS DISTINCT FROM 32 OR
       p_failure_sha256=pg_catalog.decode(pg_catalog.repeat('00',32),'hex') OR
       (p_retryable AND p_failure_class NOT IN ('transient','conflict')) OR
       (p_failure_class='cancelled' AND p_retryable) OR
       (p_provider_reference IS NOT NULL AND
        pg_catalog.octet_length(p_provider_reference) NOT BETWEEN 1 AND 512) OR
       NOT pg_catalog.isfinite(p_finished_at) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export finish input is invalid';
    END IF;
    v_digest:=pg_catalog.encode(p_failure_sha256,'hex');
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
      AND job.effect_idempotency_key=job.dedupe_key FOR UPDATE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
      AND attempt.attempt=p_attempt_generation
      AND attempt.lease_generation=p_attempt_generation
      AND attempt.lease_owner_id=p_lease_owner_id FOR UPDATE;
    IF p_failure_class='cancelled' AND v_job.cancel_requested_at IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export cancellation was not requested';
    END IF;
    IF p_failure_class='cancelled' AND
       v_attempt.effect_started_at IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export cancellation cleanup is required';
    END IF;
    IF p_failure_class='cancelled' THEN
        v_outcome:='cancelled'; v_job_state:='cancelled';
    ELSIF p_retryable AND p_attempt_generation<3 AND
          v_attempt.effect_started_at IS NULL THEN
        v_outcome:='requeued'; v_job_state:='queued';
    ELSE
        v_outcome:='dead'; v_job_state:='dead';
    END IF;
    IF v_attempt.finished_at IS NOT NULL THEN
        IF NOT COALESCE((v_job.state=v_job_state AND
           v_job.updated_at=p_finished_at AND v_job.lease_owner_id IS NULL AND
           v_job.lease_heartbeat_at IS NULL AND v_job.lease_until IS NULL AND
           v_job.last_error_class=p_failure_class AND
           v_job.last_error_code=p_failure_code AND
           v_job.last_error_sha256=v_digest AND
           v_job.effect_started_at IS NOT DISTINCT FROM
               v_attempt.effect_started_at AND
           ((v_job_state='queued' AND v_job.completed_at IS NULL AND
             v_job.retain_until IS NULL AND v_job.run_at>=p_finished_at AND
             v_job.run_at<=p_finished_at+interval '15 minutes') OR
            (v_job_state='dead' AND v_job.completed_at=p_finished_at AND
             v_job.retain_until=p_finished_at+interval '366 days') OR
            (v_job_state='cancelled' AND v_job.completed_at=p_finished_at AND
             v_job.retain_until=p_finished_at+interval '91 days')) AND
           v_attempt.outcome=v_outcome AND
           v_attempt.error_class=p_failure_class AND
           v_attempt.error_code=p_failure_code AND
           v_attempt.redacted_error_sha256=v_digest AND
           v_attempt.finished_at=p_finished_at),false) THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export finish replay differs';
        END IF;
        IF v_attempt.effect_started_at IS NOT NULL THEN
            IF v_attempt.effect_receipt_id IS NULL THEN
                RAISE EXCEPTION USING ERRCODE='P6048',
                    MESSAGE='Analytics dashboard export finish receipt is missing';
            END IF;
            PERFORM platform.platform_async_analytics_dashboard_export_effect_receipt_v1(
                p_tenant_id,p_export_id,p_job_id,p_attempt_id,
                v_attempt.effect_receipt_id,'failed',p_failure_class,
                p_failure_code,p_provider_reference,NULL,p_finished_at);
        ELSIF v_attempt.effect_receipt_id IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Analytics dashboard export finish receipt is unexpected';
        END IF;
        job_state:=v_job_state;
        effect_receipt_id:=v_attempt.effect_receipt_id;
        RETURN NEXT; RETURN;
    END IF;
    IF v_job.state IS DISTINCT FROM 'running' OR v_attempt.attempt_id IS NULL OR
       v_job.attempts IS DISTINCT FROM p_attempt_generation OR
       v_job.lease_generation IS DISTINCT FROM p_attempt_generation OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_until<=pg_catalog.clock_timestamp() OR
       p_finished_at<=v_job.updated_at THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export finish lease is stale';
    END IF;
    IF v_attempt.effect_started_at IS NOT NULL THEN
        v_receipt_id:=platform.platform_async_analytics_dashboard_export_effect_receipt_v1(
            p_tenant_id,p_export_id,p_job_id,p_attempt_id,NULL,'failed',
            p_failure_class,p_failure_code,p_provider_reference,NULL,p_finished_at);
    END IF;
    UPDATE platform.job_attempts SET finished_at=p_finished_at,
        outcome=v_outcome,error_class=p_failure_class,error_code=p_failure_code,
        redacted_error_sha256=v_digest,effect_receipt_id=v_receipt_id
    WHERE attempt_id=p_attempt_id;
    v_retry_at:=CASE WHEN v_job_state='queued' THEN p_finished_at+
        pg_catalog.random()*LEAST(
            interval '5 seconds'*pg_catalog.power(
                2::double precision,(p_attempt_generation-1)::double precision),
            interval '15 minutes') ELSE p_finished_at END;
    UPDATE platform.jobs SET state=v_job_state,
        run_at=CASE WHEN v_job_state='queued' THEN v_retry_at ELSE run_at END,
        lease_owner_id=NULL,lease_heartbeat_at=NULL,lease_until=NULL,
        last_error_class=p_failure_class,last_error_code=p_failure_code,
        last_error_sha256=v_digest,
        retain_until=CASE WHEN v_job_state='dead' THEN
            p_finished_at+interval '366 days' WHEN v_job_state='cancelled' THEN
            p_finished_at+interval '91 days' END,updated_at=p_finished_at,
        completed_at=CASE WHEN v_job_state='queued' THEN NULL ELSE p_finished_at END
    WHERE id=p_job_id;
    UPDATE platform.artifact_quota_counters
    SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
        revision=revision+1,updated_at=GREATEST(
            updated_at+interval '1 microsecond',p_finished_at)
    WHERE tenant_id=p_tenant_id AND reserved_ciphertext_bytes>=2147483648;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6049',
            MESSAGE='Analytics dashboard export finish quota is inconsistent';
    END IF;
    job_state:=v_job_state; effect_receipt_id:=v_receipt_id;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_finish_v1(
        uuid,uuid,uuid,uuid,bigint,uuid,text,text,bytea,boolean,text,timestamptz)
FROM PUBLIC;

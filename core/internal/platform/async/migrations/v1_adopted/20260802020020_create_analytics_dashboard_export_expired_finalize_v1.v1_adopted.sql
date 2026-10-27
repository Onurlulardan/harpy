-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_analytics_dashboard_export_expired_finalize_v1(
    p_tenant_id uuid,p_export_id uuid,p_job_id uuid,p_attempt_id uuid,
    p_attempt_generation bigint,p_lease_owner_id uuid,p_effect_id uuid,
    p_request_sha256 bytea,p_observed_at timestamptz
)
RETURNS TABLE(final_state text,finalized_at timestamptz,failure_sha256 bytea)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_request text;
    v_error_class text;
    v_error_code text;
    v_error_digest text;
    v_outcome text;
    v_job_state text;
    v_finalized_at timestamptz;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Analytics dashboard export expiry authority is invalid';
    END IF;
    IF p_tenant_id IS NULL OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_lease_owner_id) IS DISTINCT FROM 7 OR
       NOT pg_catalog.isfinite(p_observed_at) OR
       p_observed_at>pg_catalog.clock_timestamp()+interval '5 seconds' OR NOT (
        p_effect_id IS NULL AND p_request_sha256 IS NULL OR
        pg_catalog.uuid_extract_version(p_effect_id)=7 AND
        pg_catalog.octet_length(p_request_sha256)=32 AND
        p_request_sha256<>pg_catalog.decode(pg_catalog.repeat('00',32),'hex')
       ) THEN
        RAISE EXCEPTION USING ERRCODE='22023',
            MESSAGE='Analytics dashboard export expiry input is invalid';
    END IF;
    v_request:=CASE WHEN p_request_sha256 IS NOT NULL THEN
        pg_catalog.encode(p_request_sha256,'hex') END;
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
    IF v_job.id IS NULL OR v_attempt.attempt_id IS NULL OR
       v_attempt.effect_id IS DISTINCT FROM p_effect_id OR
       v_attempt.effect_request_sha256 IS DISTINCT FROM v_request THEN
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export expiry evidence differs';
    END IF;
    PERFORM platform.platform_async_analytics_dashboard_export_assert_cleanup_owner_v1(
        v_job.scope_kind,v_job.scope_id,v_job.authority_group_id,
        v_job.authority_domain);
    IF v_attempt.finished_at IS NOT NULL THEN
        IF v_job.state='dead' AND v_job.last_error_class IS NULL AND
           v_job.last_error_code='JOB_OUTCOME_UNKNOWN' AND
           v_attempt.outcome='dead' AND v_attempt.error_class IS NULL AND
           v_attempt.error_code='JOB_OUTCOME_UNKNOWN' AND
           v_attempt.effect_started_at IS NOT NULL AND
           v_attempt.finished_at=v_job.completed_at THEN
            final_state:='outcome_unknown'; finalized_at:=v_job.completed_at;
            RETURN NEXT; RETURN;
        END IF;
        IF v_attempt.effect_started_at IS NULL AND
           v_attempt.finished_at=v_job.updated_at AND (
            v_job.state='queued' AND v_attempt.outcome='requeued' OR
            v_job.state='dead' AND v_attempt.outcome='dead' OR
            v_job.state='cancelled' AND v_attempt.outcome='cancelled') AND
           v_attempt.error_class=v_job.last_error_class AND
           v_attempt.error_code=v_job.last_error_code AND
           v_attempt.redacted_error_sha256=v_job.last_error_sha256 THEN
            final_state:=v_job.state; finalized_at:=v_attempt.finished_at;
            failure_sha256:=pg_catalog.decode(v_attempt.redacted_error_sha256,'hex');
            RETURN NEXT; RETURN;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Analytics dashboard export expiry replay differs';
    END IF;
    IF v_job.state IS DISTINCT FROM 'running' OR
       v_job.attempts<>p_attempt_generation OR
       v_job.lease_generation<>p_attempt_generation OR
       v_job.lease_owner_id<>p_lease_owner_id OR
       v_job.lease_until>pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE='P6047',
            MESSAGE='Analytics dashboard export expiry lease is not expired';
    END IF;
    v_finalized_at:=GREATEST(pg_catalog.clock_timestamp(),p_observed_at,
        v_job.updated_at+interval '1 microsecond');
    IF v_attempt.effect_started_at IS NOT NULL THEN
        v_outcome:='dead'; v_job_state:='dead';
        UPDATE platform.job_attempts SET finished_at=v_finalized_at,
            outcome='dead',error_class=NULL,error_code='JOB_OUTCOME_UNKNOWN',
            redacted_error_sha256=NULL,effect_receipt_id=NULL
        WHERE attempt_id=p_attempt_id;
        UPDATE platform.jobs SET state='dead',lease_owner_id=NULL,
            lease_heartbeat_at=NULL,lease_until=NULL,last_error_class=NULL,
            last_error_code='JOB_OUTCOME_UNKNOWN',last_error_sha256=NULL,
            retain_until=v_finalized_at+interval '366 days',
            updated_at=v_finalized_at,completed_at=v_finalized_at
        WHERE id=p_job_id;
        final_state:='outcome_unknown'; finalized_at:=v_finalized_at;
        RETURN NEXT; RETURN;
    END IF;
    IF v_job.cancel_requested_at IS NOT NULL THEN
        v_outcome:='cancelled'; v_job_state:='cancelled';
        v_error_class:='cancelled'; v_error_code:='EXPORT_CANCELLED';
    ELSIF p_attempt_generation<3 THEN
        v_outcome:='requeued'; v_job_state:='queued';
        v_error_class:='transient'; v_error_code:='JOB_LEASE_EXPIRED';
    ELSE
        v_outcome:='dead'; v_job_state:='dead';
        v_error_class:='transient'; v_error_code:='JOB_LEASE_EXPIRED';
    END IF;
    v_error_digest:=pg_catalog.encode(pg_catalog.sha256(
        pg_catalog.convert_to('analytics.dashboard.export.expiry.v1','UTF8')||
        pg_catalog.uuid_send(p_job_id)||pg_catalog.uuid_send(p_attempt_id)||
        pg_catalog.int8send(p_attempt_generation)||
        pg_catalog.convert_to(v_error_code,'UTF8')),'hex');
    UPDATE platform.job_attempts SET finished_at=v_finalized_at,
        outcome=v_outcome,error_class=v_error_class,error_code=v_error_code,
        redacted_error_sha256=v_error_digest WHERE attempt_id=p_attempt_id;
    UPDATE platform.jobs SET state=v_job_state,
        run_at=CASE WHEN v_job_state='queued' THEN
            v_finalized_at+interval '5 seconds' ELSE run_at END,
        lease_owner_id=NULL,lease_heartbeat_at=NULL,lease_until=NULL,
        last_error_class=v_error_class,last_error_code=v_error_code,
        last_error_sha256=v_error_digest,
        retain_until=CASE WHEN v_job_state='dead' THEN v_finalized_at+interval '366 days'
            WHEN v_job_state='cancelled' THEN v_finalized_at+interval '91 days' END,
        updated_at=v_finalized_at,
        completed_at=CASE WHEN v_job_state='queued' THEN NULL ELSE v_finalized_at END
    WHERE id=p_job_id;
    UPDATE platform.artifact_quota_counters
    SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
        revision=revision+1,updated_at=GREATEST(
            updated_at+interval '1 microsecond',v_finalized_at)
    WHERE tenant_id=p_tenant_id AND reserved_ciphertext_bytes>=2147483648;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='P6049',
            MESSAGE='Analytics dashboard export expiry quota is inconsistent';
    END IF;
    final_state:=v_job_state; finalized_at:=v_finalized_at;
    failure_sha256:=pg_catalog.decode(v_error_digest,'hex'); RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_analytics_dashboard_export_expired_finalize_v1(
        uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,timestamptz)
FROM PUBLIC;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_export_finish_v1(
    p_tenant_id uuid,
    p_export_id uuid,
    p_job_id uuid,
    p_attempt_id uuid,
    p_attempt_generation bigint,
    p_lease_owner_id uuid,
    p_failure_class text,
    p_failure_code text,
    p_failure_sha256 bytea,
    p_retryable boolean,
    p_provider_reference text,
    p_finished_at timestamptz
)
RETURNS text
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
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
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       p_failure_class NOT IN ('transient','permanent','conflict','cancelled') OR
       p_failure_code !~ '^[A-Z][A-Z0-9_]{1,63}$' OR
       pg_catalog.octet_length(p_failure_sha256) IS DISTINCT FROM 32 OR
       (p_retryable AND p_failure_class NOT IN ('transient','conflict')) OR
       (p_failure_class='cancelled' AND p_retryable) OR
       NOT pg_catalog.isfinite(p_finished_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Forms export finish input is invalid';
    END IF;
    v_digest := pg_catalog.encode(p_failure_sha256,'hex');
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.kind='forms.record-export' AND job.payload_revision=1
      AND job.authority_domain='forms' AND job.writer_origin='v2'
      AND job.payload_ref='forms-record-export/' || p_export_id::text
      AND job.max_attempts=3 FOR UPDATE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.attempt=p_attempt_generation
      AND attempt.lease_generation=p_attempt_generation
      AND attempt.lease_owner_id=p_lease_owner_id FOR UPDATE;
    IF p_failure_class='cancelled' THEN
        v_outcome := 'cancelled'; v_job_state := 'cancelled';
    ELSIF p_retryable AND p_attempt_generation < 3 THEN
        v_outcome := 'requeued'; v_job_state := 'queued';
    ELSE
        v_outcome := 'dead'; v_job_state := 'dead';
    END IF;
    IF v_attempt.finished_at IS NOT NULL THEN
        IF v_job.state IS DISTINCT FROM v_job_state OR
           v_attempt.outcome IS DISTINCT FROM v_outcome OR
           v_attempt.error_class IS DISTINCT FROM p_failure_class OR
           v_attempt.error_code IS DISTINCT FROM p_failure_code OR
           v_attempt.redacted_error_sha256 IS DISTINCT FROM v_digest OR
           v_attempt.finished_at IS DISTINCT FROM p_finished_at THEN
            RAISE EXCEPTION USING ERRCODE='P6048',
                MESSAGE='Forms export finish replay differs';
        END IF;
        IF v_attempt.effect_started_at IS NOT NULL THEN
            IF v_attempt.effect_receipt_id IS NULL THEN
                RAISE EXCEPTION USING ERRCODE='P6048',
                    MESSAGE='Forms export finish receipt is missing';
            END IF;
            PERFORM platform.platform_async_forms_export_effect_receipt_v1(
                p_tenant_id,p_export_id,p_job_id,p_attempt_id,
                v_attempt.effect_receipt_id,'failed',p_failure_class,
                p_failure_code,p_provider_reference,NULL,p_finished_at
            );
        END IF;
        RETURN v_job_state;
    END IF;
    IF v_job.state IS DISTINCT FROM 'running' OR v_attempt.attempt_id IS NULL OR
       v_attempt.finished_at IS NOT NULL OR
       v_job.attempts IS DISTINCT FROM p_attempt_generation OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_until <= pg_catalog.clock_timestamp() OR
       p_finished_at <= v_job.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6047',
            MESSAGE = 'Forms export finish lease is stale';
    END IF;
    IF v_attempt.effect_started_at IS NOT NULL THEN
        v_receipt_id := platform.platform_async_forms_export_effect_receipt_v1(
            p_tenant_id,p_export_id,p_job_id,p_attempt_id,NULL,'failed',
            p_failure_class,p_failure_code,p_provider_reference,NULL,p_finished_at
        );
    END IF;
    UPDATE platform.job_attempts SET finished_at=p_finished_at,
        outcome=v_outcome,error_class=p_failure_class,error_code=p_failure_code,
        redacted_error_sha256=v_digest,effect_receipt_id=v_receipt_id
    WHERE attempt_id=p_attempt_id;
    v_retry_at := CASE WHEN v_job_state='queued' THEN p_finished_at +
        pg_catalog.random() * LEAST(
            interval '5 seconds' * pg_catalog.power(
                2::double precision,(p_attempt_generation-1)::double precision
            ),interval '15 minutes'
        ) ELSE p_finished_at END;
    UPDATE platform.jobs SET state=v_job_state,
        run_at=CASE WHEN v_job_state='queued'
                    THEN v_retry_at ELSE run_at END,
        lease_owner_id=NULL,lease_heartbeat_at=NULL,lease_until=NULL,
        last_error_class=p_failure_class,last_error_code=p_failure_code,
        last_error_sha256=v_digest,
        retain_until=CASE WHEN v_job_state='dead'
                     THEN p_finished_at+interval '366 days'
                     WHEN v_job_state='cancelled'
                     THEN p_finished_at+interval '91 days' END,
        updated_at=p_finished_at,
        completed_at=CASE WHEN v_job_state='queued' THEN NULL ELSE p_finished_at END
    WHERE id=p_job_id;
    UPDATE platform.artifact_quota_counters
    SET reserved_ciphertext_bytes=reserved_ciphertext_bytes-2147483648,
        revision=revision+1,updated_at=p_finished_at
    WHERE tenant_id=p_tenant_id AND reserved_ciphertext_bytes>=2147483648;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6049',
            MESSAGE = 'Forms export finish quota is inconsistent';
    END IF;
    RETURN v_job_state;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_finish_v1(
    uuid,uuid,uuid,uuid,bigint,uuid,text,text,bytea,boolean,text,timestamptz
) FROM PUBLIC;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_export_unknown_v1(
    p_tenant_id uuid,
    p_export_id uuid,
    p_job_id uuid,
    p_attempt_id uuid,
    p_attempt_generation bigint,
    p_lease_owner_id uuid,
    p_effect_id uuid,
    p_request_sha256 bytea,
    p_observed_at timestamptz
)
RETURNS boolean
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE = '22023',
            MESSAGE = 'Forms export unknown input is invalid';
    END IF;
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
    IF v_attempt.finished_at IS NOT NULL THEN
        IF v_job.state='dead' AND v_job.last_error_class IS NULL AND
           v_job.last_error_code='JOB_OUTCOME_UNKNOWN' AND
           v_job.completed_at=p_observed_at AND v_attempt.outcome='dead' AND
           v_attempt.error_class IS NULL AND
           v_attempt.error_code='JOB_OUTCOME_UNKNOWN' AND
           v_attempt.redacted_error_sha256 IS NULL AND
           v_attempt.effect_receipt_id IS NULL AND
           v_attempt.effect_id=p_effect_id AND
           v_attempt.effect_request_sha256=
               pg_catalog.encode(p_request_sha256,'hex') AND
           v_attempt.finished_at=p_observed_at THEN
            RETURN true;
        END IF;
        RAISE EXCEPTION USING ERRCODE='P6048',
            MESSAGE='Forms export unknown replay differs';
    END IF;
    IF v_job.state IS DISTINCT FROM 'running' OR v_attempt.attempt_id IS NULL OR
       v_attempt.finished_at IS NOT NULL OR
       v_job.attempts IS DISTINCT FROM p_attempt_generation OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_until <= pg_catalog.clock_timestamp() OR
       v_attempt.effect_id IS DISTINCT FROM p_effect_id OR
       v_attempt.effect_request_sha256 IS DISTINCT FROM
           pg_catalog.encode(p_request_sha256,'hex') OR
       v_attempt.effect_started_at IS NULL OR
       v_job.effect_started_at IS DISTINCT FROM v_attempt.effect_started_at OR
       p_observed_at <= v_job.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6047',
            MESSAGE = 'Forms export unknown lease is stale';
    END IF;
    UPDATE platform.job_attempts SET finished_at=p_observed_at,outcome='dead',
        error_class=NULL,error_code='JOB_OUTCOME_UNKNOWN',
        redacted_error_sha256=NULL,effect_receipt_id=NULL
    WHERE attempt_id=p_attempt_id;
    UPDATE platform.jobs SET state='dead',lease_owner_id=NULL,
        lease_heartbeat_at=NULL,lease_until=NULL,last_error_class=NULL,
        last_error_code='JOB_OUTCOME_UNKNOWN',last_error_sha256=NULL,
        retain_until=p_observed_at+interval '366 days',
        updated_at=p_observed_at,completed_at=p_observed_at
    WHERE id=p_job_id;
    RETURN true;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_unknown_v1(
    uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,timestamptz
) FROM PUBLIC;
-- +goose StatementEnd

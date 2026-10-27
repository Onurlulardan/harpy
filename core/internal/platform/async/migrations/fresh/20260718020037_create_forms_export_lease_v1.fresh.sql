-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_export_checkpoint_v1(
    p_tenant_id uuid,
    p_export_id uuid,
    p_job_id uuid,
    p_attempt_id uuid,
    p_attempt_generation bigint,
    p_lease_owner_id uuid,
    p_observed_at timestamptz
)
RETURNS TABLE(lease_until timestamptz, cancellation_requested boolean)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR
       p_lease_owner_id IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms export checkpoint authority is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.kind='forms.record-export' AND job.payload_revision=1
      AND job.authority_domain='forms' AND job.writer_origin='v2'
      AND job.payload_ref='forms-record-export/' || p_export_id::text
      AND job.max_attempts=3
    FOR UPDATE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.scope_kind='tenant' AND attempt.scope_id=p_tenant_id
      AND attempt.attempt=p_attempt_generation
      AND attempt.lease_generation=p_attempt_generation
      AND attempt.lease_owner_id=p_lease_owner_id
    FOR UPDATE;
    IF v_job.state IS DISTINCT FROM 'running' OR v_attempt.attempt_id IS NULL OR
       v_attempt.finished_at IS NOT NULL OR
       v_job.attempts IS DISTINCT FROM p_attempt_generation OR
       v_job.lease_generation IS DISTINCT FROM p_attempt_generation OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_until <= pg_catalog.clock_timestamp() OR
       p_observed_at <= v_job.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6047',
            MESSAGE = 'Forms export lease is stale';
    END IF;
    UPDATE platform.jobs SET lease_heartbeat_at=p_observed_at,
        lease_until=p_observed_at + interval '60 seconds',updated_at=p_observed_at
    WHERE id=p_job_id;
    lease_until := p_observed_at + interval '60 seconds';
    cancellation_requested := v_job.cancel_requested_at IS NOT NULL;
    RETURN NEXT;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_checkpoint_v1(
    uuid,uuid,uuid,uuid,bigint,uuid,timestamptz
) FROM PUBLIC;

CREATE FUNCTION platform.platform_async_forms_export_effect_start_v1(
    p_tenant_id uuid,
    p_export_id uuid,
    p_job_id uuid,
    p_attempt_id uuid,
    p_attempt_generation bigint,
    p_lease_owner_id uuid,
    p_effect_id uuid,
    p_request_sha256 bytea,
    p_started_at timestamptz
)
RETURNS timestamptz
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_attempt platform.job_attempts%ROWTYPE;
    v_request text;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_job_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       pg_catalog.uuid_extract_version(p_effect_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR p_lease_owner_id IS NULL OR
       pg_catalog.octet_length(p_request_sha256) IS DISTINCT FROM 32 OR
       NOT pg_catalog.isfinite(p_started_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms export effect authority is invalid';
    END IF;
    v_request := pg_catalog.encode(p_request_sha256,'hex');
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.id=p_job_id AND job.tenant_id=p_tenant_id
      AND job.scope_kind='tenant' AND job.scope_id=p_tenant_id
      AND job.kind='forms.record-export' AND job.payload_revision=1
      AND job.authority_domain='forms' AND job.writer_origin='v2'
      AND job.payload_ref='forms-record-export/' || p_export_id::text
      AND job.effect_class='external' AND job.max_attempts=3
    FOR UPDATE;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id=p_attempt_id AND attempt.job_id=p_job_id
      AND attempt.attempt=p_attempt_generation
      AND attempt.lease_generation=p_attempt_generation
      AND attempt.lease_owner_id=p_lease_owner_id
    FOR UPDATE;
    IF v_attempt.effect_started_at IS NOT NULL THEN
        IF v_attempt.effect_id=p_effect_id AND
           v_attempt.effect_request_sha256=v_request AND
           v_job.effect_started_at=v_attempt.effect_started_at THEN
            RETURN v_attempt.effect_started_at;
        END IF;
        RAISE EXCEPTION USING ERRCODE = 'P6048',
            MESSAGE = 'Forms export effect replay differs';
    END IF;
    IF v_job.state IS DISTINCT FROM 'running' OR v_attempt.attempt_id IS NULL OR
       v_attempt.finished_at IS NOT NULL OR v_job.cancel_requested_at IS NOT NULL OR
       v_job.attempts IS DISTINCT FROM p_attempt_generation OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_until <= pg_catalog.clock_timestamp() OR
       p_started_at <= v_job.updated_at OR p_started_at < v_attempt.started_at THEN
        RAISE EXCEPTION USING ERRCODE = 'P6047',
            MESSAGE = 'Forms export effect lease is stale';
    END IF;
    UPDATE platform.jobs SET effect_started_at=p_started_at,updated_at=p_started_at
    WHERE id=p_job_id;
    UPDATE platform.job_attempts
    SET effect_started_at=p_started_at,effect_id=p_effect_id,
        effect_request_sha256=v_request
    WHERE attempt_id=p_attempt_id;
    RETURN p_started_at;
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_effect_start_v1(
    uuid,uuid,uuid,uuid,bigint,uuid,uuid,bytea,timestamptz
) FROM PUBLIC;
-- +goose StatementEnd

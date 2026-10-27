-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_import_complete_v1(
    p_tenant_id uuid,
    p_import_id uuid,
    p_job_id uuid,
    p_attempt_id uuid,
    p_lease_owner_id uuid,
    p_lease_generation bigint,
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
    v_transition_at timestamptz;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_lease_owner_id IS NULL OR p_lease_generation < 1 OR
       NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms import completion authority is invalid';
    END IF;
    SELECT job.* INTO v_job
    FROM platform.jobs AS job
    WHERE job.id = p_job_id
      AND job.tenant_id = p_tenant_id
      AND job.scope_kind = 'tenant'
      AND job.scope_id = p_tenant_id
      AND job.authority_domain = 'forms'
      AND job.writer_origin = 'v2'
      AND job.kind = 'forms.record-import'
      AND job.payload_revision = 1
      AND job.effect_class = 'database_transactional'
      AND job.effect_idempotency_key IS NULL
      AND job.dedupe_key = 'forms.record-import:' || p_import_id::text
      AND job.payload_ref = 'forms-record-import/' || p_import_id::text
      AND job.max_attempts = 8
    FOR UPDATE;
    IF NOT FOUND OR v_job.state IS DISTINCT FROM 'running' OR
       v_job.cancel_requested_at IS NOT NULL OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_generation IS DISTINCT FROM p_lease_generation OR
       v_job.lease_until <= pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE = 'P6046',
            MESSAGE = 'Forms import completion lease is stale';
    END IF;
    SELECT attempt.* INTO v_attempt
    FROM platform.job_attempts AS attempt
    WHERE attempt.attempt_id = p_attempt_id
      AND attempt.job_id = p_job_id
      AND attempt.scope_kind = 'tenant'
      AND attempt.scope_id = p_tenant_id
      AND attempt.attempt = p_lease_generation
      AND attempt.lease_generation = p_lease_generation
      AND attempt.lease_owner_id = p_lease_owner_id
    FOR UPDATE;
    IF NOT FOUND OR v_attempt.finished_at IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P6046',
            MESSAGE = 'Forms import completion attempt is stale';
    END IF;
    v_transition_at := GREATEST(
        p_observed_at,v_job.updated_at + interval '1 microsecond',
        v_attempt.started_at
    );
    UPDATE platform.job_attempts
    SET finished_at=v_transition_at,outcome='succeeded'
    WHERE attempt_id=p_attempt_id;
    UPDATE platform.jobs
    SET state='succeeded',lease_owner_id=NULL,lease_heartbeat_at=NULL,
        lease_until=NULL,last_error_class=NULL,last_error_code=NULL,
        last_error_sha256=NULL,retain_until=v_transition_at + interval '91 days',
        updated_at=v_transition_at,completed_at=v_transition_at
    WHERE id=p_job_id;
    RETURN true;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_forms_import_complete_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,timestamptz
) FROM PUBLIC;

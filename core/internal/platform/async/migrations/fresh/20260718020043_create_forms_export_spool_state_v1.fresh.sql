-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_export_spool_state_v1(
    p_tenant_id uuid,
    p_export_id uuid,
    p_attempt_generation bigint,
    p_lease_owner_id uuid
)
RETURNS text
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
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 OR
       p_attempt_generation NOT BETWEEN 1 AND 3 OR p_lease_owner_id IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms export spool state authority is invalid';
    END IF;
    SELECT job.* INTO v_job FROM platform.jobs AS job
    WHERE job.tenant_id=p_tenant_id AND job.scope_kind='tenant'
      AND job.scope_id=p_tenant_id AND job.kind='forms.record-export'
      AND job.payload_revision=1 AND job.authority_domain='forms'
      AND job.writer_origin='v2'
      AND job.payload_ref='forms-record-export/' || p_export_id::text
      AND job.max_attempts=3;
    IF NOT FOUND OR v_job.attempts<>p_attempt_generation OR
       v_job.lease_generation<>p_attempt_generation THEN
        RETURN 'inactive';
    END IF;
    SELECT attempt.* INTO v_attempt FROM platform.job_attempts AS attempt
    WHERE attempt.job_id=v_job.id AND attempt.attempt=p_attempt_generation
      AND attempt.lease_generation=p_attempt_generation
      AND attempt.lease_owner_id=p_lease_owner_id;
    IF NOT FOUND THEN RETURN 'inactive'; END IF;
    IF v_job.state='running' AND v_attempt.finished_at IS NULL AND
       v_job.lease_owner_id=p_lease_owner_id AND
       v_job.lease_until>pg_catalog.clock_timestamp() THEN
        RETURN 'running';
    END IF;
    IF v_job.state='dead' AND v_job.last_error_class IS NULL AND
       v_job.last_error_code='JOB_OUTCOME_UNKNOWN' AND
       v_attempt.outcome='dead' AND v_attempt.error_class IS NULL AND
       v_attempt.error_code='JOB_OUTCOME_UNKNOWN' THEN
        RETURN 'outcome_unknown';
    END IF;
    RETURN 'inactive';
END;
$function$;

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_spool_state_v1(
    uuid,uuid,bigint,uuid
) FROM PUBLIC;

CREATE FUNCTION platform.platform_async_forms_export_spool_tenant_v1(
    p_export_id uuid
)
RETURNS uuid
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE v_tenant_id uuid;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       pg_catalog.uuid_extract_version(p_export_id) IS DISTINCT FROM 7 THEN
        RAISE EXCEPTION USING ERRCODE='42501',
            MESSAGE='Forms export spool tenant authority is invalid';
    END IF;
    SELECT job.tenant_id INTO v_tenant_id FROM platform.jobs AS job
    WHERE job.scope_kind='tenant' AND job.scope_id=job.tenant_id
      AND job.authority_domain='forms' AND job.writer_origin='v2'
      AND job.kind='forms.record-export' AND job.payload_revision=1
      AND job.payload_ref='forms-record-export/' || p_export_id::text;
    RETURN v_tenant_id;
END;
$function$;

REVOKE ALL ON FUNCTION
    platform.platform_async_forms_export_spool_tenant_v1(uuid) FROM PUBLIC;
-- +goose StatementEnd

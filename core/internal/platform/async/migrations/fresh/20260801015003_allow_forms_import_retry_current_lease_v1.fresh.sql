-- +goose Up
-- +goose StatementBegin
DROP FUNCTION platform.platform_async_forms_import_current_lease_v1(
    uuid,uuid,uuid,uuid,uuid,bigint
);
CREATE FUNCTION platform.platform_async_forms_import_current_lease_v1(
    p_tenant_id uuid,
    p_import_id uuid,
    p_job_id uuid,
    p_attempt_id uuid,
    p_lease_owner_id uuid,
    p_lease_generation bigint
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
    IF CURRENT_USER IS DISTINCT FROM 'harpy_owner' OR
       SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       pg_catalog.uuid_extract_version(p_attempt_id) IS DISTINCT FROM 7 OR
       p_lease_owner_id IS NULL OR p_lease_generation < 1 THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms import lease authority is invalid';
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
      AND ((job.retry_of_job_id IS NULL AND job.dedupe_key =
            'forms.record-import:' || p_import_id::text) OR
           (job.retry_of_job_id IS NOT NULL AND job.dedupe_key =
            'forms.record-import.retry:' || job.retry_of_job_id::text))
      AND job.payload_ref = 'forms-record-import/' || p_import_id::text
      AND job.max_attempts = 8
    FOR SHARE;
    IF NOT FOUND OR v_job.state IS DISTINCT FROM 'running' OR
       v_job.lease_owner_id IS DISTINCT FROM p_lease_owner_id OR
       v_job.lease_generation IS DISTINCT FROM p_lease_generation OR
       v_job.lease_until <= pg_catalog.clock_timestamp() THEN
        RAISE EXCEPTION USING ERRCODE = 'P6046',
            MESSAGE = 'Forms import lease is stale';
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
    FOR SHARE;
    IF NOT FOUND OR v_attempt.finished_at IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P6046',
            MESSAGE = 'Forms import attempt is stale';
    END IF;
    RETURN v_job.cancel_requested_at IS NOT NULL;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_forms_import_current_lease_v1(
    uuid,uuid,uuid,uuid,uuid,bigint
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_import_current_lease_v1(
    uuid,uuid,uuid,uuid,uuid,bigint
) TO harpy_runtime;

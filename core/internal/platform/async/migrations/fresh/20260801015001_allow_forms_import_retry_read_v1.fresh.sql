-- +goose Up
-- +goose StatementBegin
DROP FUNCTION platform.platform_async_forms_import_read_v1(uuid,uuid,uuid);
CREATE FUNCTION platform.platform_async_forms_import_read_v1(
    p_tenant_id uuid,
    p_import_id uuid,
    p_job_id uuid
)
RETURNS TABLE(
    job_state text,
    cancellation_requested boolean,
    attempts integer,
    max_attempts integer,
    completed_at timestamptz
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms import job read authority is invalid';
    END IF;
    RETURN QUERY
    SELECT job.state::text,job.cancel_requested_at IS NOT NULL,
           job.attempts,job.max_attempts,job.completed_at
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
      AND (job.payload_ref = 'forms-record-import/' || p_import_id::text OR
           (job.payload_ref IS NULL AND job.payload_redacted_at IS NOT NULL))
      AND job.max_attempts = 8;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6044',
            MESSAGE = 'Forms import job was not found';
    END IF;
END;
$function$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION platform.platform_async_forms_import_read_v1(
    uuid,uuid,uuid
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_import_read_v1(
    uuid,uuid,uuid
) TO forgecrud_runtime;

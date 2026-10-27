-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_forms_export_control_v1(
    p_tenant_id uuid,
    p_export_id uuid,
    p_job_id uuid,
    p_allow_cancel boolean,
    p_observed_at timestamptz
)
RETURNS TABLE(
    job_state text,
    cancellation_requested boolean,
    completed_at timestamptz,
    changed boolean
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_job platform.jobs%ROWTYPE;
    v_transition_at timestamptz;
BEGIN
    IF CURRENT_USER IS DISTINCT FROM 'forgecrud_owner' OR
       SESSION_USER IS DISTINCT FROM 'forgecrud_v2_runtime' OR
       p_allow_cancel IS NULL OR NOT pg_catalog.isfinite(p_observed_at) THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'Forms export job control authority is invalid';
    END IF;
    SELECT job.* INTO v_job
    FROM platform.jobs AS job
    WHERE job.id = p_job_id
      AND job.tenant_id = p_tenant_id
      AND job.scope_kind = 'tenant'
      AND job.scope_id = p_tenant_id
      AND job.authority_domain = 'forms'
      AND job.writer_origin = 'v2'
      AND job.kind = 'forms.record-export'
      AND job.payload_revision = 1
      AND job.dedupe_key = 'forms.record-export:' || p_export_id::text
      AND job.effect_idempotency_key = 'forms.record-export:' || p_export_id::text
      AND (job.payload_ref = 'forms-record-export/' || p_export_id::text OR
           (job.payload_ref IS NULL AND job.payload_redacted_at IS NOT NULL))
      AND job.max_attempts = 3
    FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE = 'P6034',
            MESSAGE = 'Forms export job was not found';
    END IF;
    changed := false;
    IF p_allow_cancel AND v_job.state = 'queued' THEN
        v_transition_at := GREATEST(
            p_observed_at, v_job.updated_at + interval '1 microsecond'
        );
        UPDATE platform.jobs
        SET state = 'cancelled', last_error_class = 'cancelled',
            last_error_code = 'EXPORT_CANCELLED',
            last_error_sha256 = pg_catalog.encode(pg_catalog.sha256(
                pg_catalog.convert_to('EXPORT_CANCELLED', 'UTF8')), 'hex'),
            retain_until = v_transition_at + interval '91 days',
            updated_at = v_transition_at, completed_at = v_transition_at
        WHERE id = v_job.id
        RETURNING * INTO v_job;
        changed := true;
    ELSIF p_allow_cancel AND v_job.state = 'running' AND
          v_job.cancel_requested_at IS NULL THEN
        v_transition_at := GREATEST(
            p_observed_at, v_job.updated_at + interval '1 microsecond'
        );
        UPDATE platform.jobs
        SET cancel_requested_at = v_transition_at,
            updated_at = v_transition_at
        WHERE id = v_job.id
        RETURNING * INTO v_job;
        changed := true;
    END IF;
    job_state := v_job.state;
    cancellation_requested := v_job.cancel_requested_at IS NOT NULL;
    completed_at := v_job.completed_at;
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_forms_export_control_v1(
    uuid,uuid,uuid,boolean,timestamptz
) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_submit_v1(
    uuid,uuid,uuid,uuid,uuid,bigint,text,text,timestamptz
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_read_v1(
    uuid,uuid,uuid
) TO forgecrud_runtime;
GRANT EXECUTE ON FUNCTION platform.platform_async_forms_export_control_v1(
    uuid,uuid,uuid,boolean,timestamptz
) TO forgecrud_runtime;

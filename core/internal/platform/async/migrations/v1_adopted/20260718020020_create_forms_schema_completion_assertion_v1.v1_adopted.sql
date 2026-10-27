-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_assert_forms_schema_completion_job_v1(
    p_job platform.jobs
)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
DECLARE
    v_schema_job platform.schema_jobs%ROWTYPE;
    v_payload_ref text;
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_migrator'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner'
       OR p_job.id IS NULL
       OR pg_catalog.uuid_extract_version(p_job.id) IS DISTINCT FROM 7
       OR p_job.payload_ref IS NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5605',
            MESSAGE = 'Forms schema completion insert authority is invalid';
    END IF;

    SELECT candidate.* INTO STRICT v_schema_job
    FROM platform.schema_jobs AS candidate
    WHERE p_job.payload_ref =
        'forms-publication-apply-result/' || candidate.job_id::text
      AND candidate.owner = 'forms'
      AND candidate.resource_kind = 'form_storage'
      AND candidate.plan_contract_version = 'forms-schema-plan.v2'
      AND candidate.compiler_version = '2.0.0'
      AND candidate.state IN ('succeeded', 'failed', 'cancelled')
      AND candidate.finished_at IS NOT NULL
    FOR SHARE;

    v_payload_ref :=
        'forms-publication-apply-result/' || v_schema_job.job_id::text;
    IF v_schema_job.submission_authority_id IS NULL
       OR v_schema_job.submission_authority_group_id IS NULL
       OR v_schema_job.submission_group_generation IS NULL
       OR v_schema_job.submission_billing_generation IS NULL
       OR v_schema_job.submission_forms_generation IS NULL
       OR p_job.tenant_id IS DISTINCT FROM v_schema_job.tenant_id
       OR p_job.scope_kind IS DISTINCT FROM 'tenant'
       OR p_job.scope_id IS DISTINCT FROM v_schema_job.tenant_id
       OR p_job.authority_group_id IS DISTINCT FROM
            v_schema_job.submission_authority_group_id
       OR p_job.authority_group_generation IS DISTINCT FROM
            v_schema_job.submission_group_generation
       OR p_job.authority_domain IS DISTINCT FROM 'forms'
       OR p_job.cutover_generation IS DISTINCT FROM
            v_schema_job.submission_forms_generation
       OR p_job.writer_origin IS DISTINCT FROM 'v2'
       OR p_job.kind IS DISTINCT FROM 'forms.publication.apply-result'
       OR p_job.payload_revision IS DISTINCT FROM 1
       OR p_job.payload_ref IS DISTINCT FROM v_payload_ref
       OR p_job.payload_sha256 IS DISTINCT FROM pg_catalog.encode(
            pg_catalog.sha256(pg_catalog.convert_to(v_payload_ref, 'UTF8')),
            'hex'
       )
       OR p_job.payload_redacted_at IS NOT NULL
       OR p_job.dedupe_key IS DISTINCT FROM
            'forms.publication.apply-result:' || v_schema_job.job_id::text
       OR p_job.effect_class IS DISTINCT FROM 'database_transactional'
       OR p_job.effect_idempotency_key IS NOT NULL
       OR p_job.priority IS DISTINCT FROM 100
       OR p_job.run_at IS DISTINCT FROM p_job.created_at
       OR p_job.state IS DISTINCT FROM 'queued'
       OR p_job.attempts IS DISTINCT FROM 0
       OR p_job.max_attempts IS DISTINCT FROM 8
       OR p_job.lease_owner_id IS NOT NULL
       OR p_job.lease_generation IS DISTINCT FROM 0
       OR p_job.lease_heartbeat_at IS NOT NULL
       OR p_job.lease_until IS NOT NULL
       OR p_job.cancel_requested_at IS NOT NULL
       OR p_job.parked_at IS NOT NULL
       OR p_job.parked_code IS NOT NULL
       OR p_job.parked_current_group_generation IS NOT NULL
       OR p_job.parked_current_cutover_generation IS NOT NULL
       OR p_job.effect_started_at IS NOT NULL
       OR p_job.last_error_class IS NOT NULL
       OR p_job.last_error_code IS NOT NULL
       OR p_job.last_error_sha256 IS NOT NULL
       OR p_job.retry_of_job_id IS NOT NULL
       OR p_job.retain_until IS NOT NULL
       OR p_job.retention_release_revision IS DISTINCT FROM 0
       OR p_job.retention_release_sha256 IS NOT NULL
       OR p_job.retention_released_at IS NOT NULL
       OR p_job.created_at IS DISTINCT FROM p_job.updated_at
       OR p_job.created_at < v_schema_job.finished_at
       OR p_job.completed_at IS NOT NULL THEN
        RAISE EXCEPTION USING ERRCODE = 'P5605',
            MESSAGE = 'Forms schema completion job differs';
    END IF;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
platform.platform_async_assert_forms_schema_completion_job_v1(platform.jobs)
FROM PUBLIC;

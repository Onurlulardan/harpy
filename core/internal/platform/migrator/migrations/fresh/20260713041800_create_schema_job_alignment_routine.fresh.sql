-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_schema_job_alignment()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM platform.schema_jobs AS job
        WHERE job.job_id = NEW.job_id AND (
            (job.state IN ('queued', 'awaiting_runtime') AND EXISTS (
                SELECT 1 FROM platform.schema_job_steps AS current_step
                WHERE current_step.job_id = job.job_id
                  AND current_step.ordinal = job.next_step_ordinal
                  AND current_step.state IN ('pending', 'retryable', 'lease_expired')
                  AND current_step.executor_kind = job.next_executor
            )) OR
            (job.state IN ('leased', 'running') AND EXISTS (
                SELECT 1 FROM platform.schema_job_steps AS current_step
                WHERE current_step.job_id = job.job_id
                  AND current_step.ordinal = job.next_step_ordinal
                  AND current_step.state = 'running'
                  AND current_step.executor_kind = job.next_executor
            )) OR
            (job.state = 'lease_expired' AND EXISTS (
                SELECT 1 FROM platform.schema_job_steps AS current_step
                WHERE current_step.job_id = job.job_id
                  AND current_step.ordinal = job.next_step_ordinal
                  AND current_step.state = 'lease_expired'
                  AND current_step.executor_kind = job.next_executor
            )) OR
            (job.state = 'succeeded' AND NOT EXISTS (
                SELECT 1 FROM platform.schema_job_steps AS unfinished_step
                WHERE unfinished_step.job_id = job.job_id
                  AND unfinished_step.state NOT IN ('succeeded', 'skipped')
            )) OR
            (job.state = 'failed' AND EXISTS (
                SELECT 1 FROM platform.schema_job_steps AS current_step
                WHERE current_step.job_id = job.job_id
                  AND current_step.ordinal = job.next_step_ordinal
                  AND current_step.state IN ('failed', 'blocked')
            )) OR
            (job.state = 'cancelled' AND EXISTS (
                SELECT 1 FROM platform.schema_job_steps AS current_step
                WHERE current_step.job_id = job.job_id
                  AND current_step.ordinal = job.next_step_ordinal
                  AND current_step.state IN (
                      'pending', 'retryable', 'lease_expired', 'cancelled'
                  )
            ))
        )
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'schema job and current step are misaligned';
    END IF;
    RETURN NULL;
END;
$function$;
-- +goose StatementEnd

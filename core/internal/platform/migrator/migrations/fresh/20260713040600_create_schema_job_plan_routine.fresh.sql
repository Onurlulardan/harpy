-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_schema_job_plan()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM platform.schema_jobs AS job
        WHERE job.job_id = NEW.job_id
          AND job.state = 'queued'
          AND job.next_executor = 'migrator'
          AND job.next_step_ordinal = 0
          AND job.generation = 1
          AND job.attempt = 0
          AND job.last_lease_expired_at IS NULL
          AND job.updated_at = job.created_at
          AND (
              SELECT pg_catalog.count(*)
              FROM platform.schema_job_steps AS step
              WHERE step.job_id = job.job_id
          ) = job.step_count
          AND (
              SELECT pg_catalog.min(step.ordinal)
              FROM platform.schema_job_steps AS step
              WHERE step.job_id = job.job_id
          ) = 0
          AND (
              SELECT pg_catalog.max(step.ordinal)
              FROM platform.schema_job_steps AS step
              WHERE step.job_id = job.job_id
          ) = job.step_count - 1
          AND EXISTS (
              SELECT 1 FROM platform.schema_job_steps AS first_step
              WHERE first_step.job_id = job.job_id AND first_step.ordinal = 0
                AND first_step.executor_kind = 'migrator'
          )
          AND NOT EXISTS (
              SELECT 1 FROM platform.schema_job_steps AS invalid_step
              WHERE invalid_step.job_id = job.job_id AND (
                  invalid_step.plan_sha256 <> job.plan_sha256 OR
                  invalid_step.state <> 'pending' OR
                  invalid_step.generation <> 1 OR invalid_step.attempt <> 0 OR
                  invalid_step.next_executor <> invalid_step.executor_kind OR
                  invalid_step.updated_at <> invalid_step.created_at
              )
          )
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'schema job plan must be complete and gapless';
    END IF;
    RETURN NULL;
END;
$function$;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_schema_attempt_insert_v2()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NEW.recorded_at > pg_catalog.clock_timestamp() OR NOT EXISTS (
        SELECT 1
        FROM platform.schema_job_steps AS step
        JOIN platform.schema_jobs AS job ON job.job_id = step.job_id
        WHERE step.job_id = NEW.job_id
          AND step.ordinal = NEW.step_ordinal
          AND step.plan_sha256 = NEW.plan_sha256
          AND step.step_sha256 = NEW.step_sha256
          AND step.state = 'running'
          AND step.attempt = NEW.attempt
          AND step.executor_kind = NEW.executor_kind
          AND step.started_at = NEW.started_at
          AND job.next_step_ordinal = NEW.step_ordinal
          AND job.next_executor = NEW.executor_kind
          AND job.lease_id = NEW.lease_id
          AND job.lease_owner_id = NEW.lease_owner_id
          AND job.state IN ('leased', 'running')
          AND (
              (NEW.state = 'lease_expired' AND NEW.error_code = 'lease_expired' AND
                  job.lease_expires_at <= NEW.finished_at) OR
              (NEW.state = 'failed' AND NEW.attempt = 8 AND
                  NEW.error_code = 'attempt_limit_exhausted' AND
                  job.lease_expires_at <= NEW.finished_at) OR
              (NEW.state <> 'lease_expired' AND
                  NEW.error_code IS DISTINCT FROM 'attempt_limit_exhausted' AND
                  job.state = 'running' AND NEW.finished_at < job.lease_expires_at)
          )
          AND COALESCE((
              SELECT pg_catalog.max(previous.attempt)
              FROM platform.schema_job_step_attempts AS previous
              WHERE previous.job_id = NEW.job_id
                AND previous.step_ordinal = NEW.step_ordinal
          ), 0) = NEW.attempt - 1
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'schema attempt does not match active lease';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

ALTER TABLE platform.schema_job_step_attempts
    DISABLE TRIGGER schema_job_step_attempts_validate_insert;

CREATE TRIGGER schema_job_step_attempts_validate_insert_v2
BEFORE INSERT ON platform.schema_job_step_attempts
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_validate_schema_attempt_insert_v2();

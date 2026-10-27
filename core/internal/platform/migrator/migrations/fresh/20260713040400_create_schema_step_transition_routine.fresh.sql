-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_schema_step_transition()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'schema step deletion is forbidden';
    END IF;
    IF ROW(
        NEW.job_id, NEW.ordinal, NEW.plan_sha256, NEW.operation_id,
        NEW.transaction_class, NEW.executor_kind, NEW.identifier_mapping,
        NEW.arguments, NEW.step_sha256, NEW.created_at
    ) IS DISTINCT FROM ROW(
        OLD.job_id, OLD.ordinal, OLD.plan_sha256, OLD.operation_id,
        OLD.transaction_class, OLD.executor_kind, OLD.identifier_mapping,
        OLD.arguments, OLD.step_sha256, OLD.created_at
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'schema step identity is immutable';
    END IF;
    IF NEW.generation <> OLD.generation + 1 OR NEW.updated_at <= OLD.updated_at THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'schema step generation must advance exactly once';
    END IF;

    IF OLD.state IN ('pending', 'retryable', 'lease_expired') AND
       NEW.state = 'running' AND NEW.attempt = OLD.attempt + 1 AND
       NEW.next_executor = OLD.executor_kind AND EXISTS (
           SELECT 1 FROM platform.schema_jobs AS job
           WHERE job.job_id = OLD.job_id AND job.state = 'leased'
             AND job.next_step_ordinal = OLD.ordinal
             AND job.next_executor = OLD.executor_kind
       ) THEN
        RETURN NEW;
    END IF;

    IF OLD.state = 'running' AND NEW.state IN (
        'succeeded', 'retryable', 'lease_expired', 'failed',
        'blocked', 'skipped', 'cancelled'
    ) AND NEW.attempt = OLD.attempt AND EXISTS (
        SELECT 1 FROM platform.schema_job_step_attempts AS receipt
        WHERE receipt.job_id = OLD.job_id
          AND receipt.step_ordinal = OLD.ordinal
          AND receipt.attempt = OLD.attempt
          AND receipt.plan_sha256 = OLD.plan_sha256
          AND receipt.step_sha256 = OLD.step_sha256
          AND receipt.state = NEW.state
          AND receipt.started_at = OLD.started_at
          AND receipt.finished_at = NEW.finished_at
    ) THEN
        RETURN NEW;
    END IF;

    RAISE EXCEPTION USING
        ERRCODE = '55000', MESSAGE = 'invalid schema step transition';
END;
$function$;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_schema_job_transition()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'schema job deletion is forbidden';
    END IF;
    IF ROW(
        NEW.job_id, NEW.tenant_id, NEW.owner, NEW.resource_kind,
        NEW.resource_id, NEW.resource_revision, NEW.compiler_version,
        NEW.plan_contract_version, NEW.canonical_plan, NEW.plan_sha256,
        NEW.step_count, NEW.created_at
    ) IS DISTINCT FROM ROW(
        OLD.job_id, OLD.tenant_id, OLD.owner, OLD.resource_kind,
        OLD.resource_id, OLD.resource_revision, OLD.compiler_version,
        OLD.plan_contract_version, OLD.canonical_plan, OLD.plan_sha256,
        OLD.step_count, OLD.created_at
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'schema job identity is immutable';
    END IF;
    IF NEW.generation <> OLD.generation + 1 OR NEW.updated_at <= OLD.updated_at THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'schema job generation must advance exactly once';
    END IF;

    IF OLD.state IN ('leased', 'running') AND NEW.state = OLD.state AND
       NEW.next_step_ordinal = OLD.next_step_ordinal AND
       NEW.next_executor = OLD.next_executor AND NEW.attempt = OLD.attempt AND
       NEW.available_at = OLD.available_at AND NEW.lease_id = OLD.lease_id AND
       NEW.lease_owner_id = OLD.lease_owner_id AND NEW.leased_at = OLD.leased_at AND
       NEW.lease_heartbeat_at > OLD.lease_heartbeat_at AND
       NEW.lease_expires_at >= OLD.lease_expires_at AND
       NEW.last_lease_expired_at IS NOT DISTINCT FROM OLD.last_lease_expired_at THEN
        RETURN NEW;
    END IF;

    IF OLD.state IN ('queued', 'awaiting_runtime') AND NEW.state = 'leased' AND
       NEW.next_step_ordinal = OLD.next_step_ordinal AND
       NEW.next_executor = OLD.next_executor AND NEW.attempt = OLD.attempt + 1 AND
       NEW.available_at = OLD.available_at AND
       NEW.last_lease_expired_at IS NOT DISTINCT FROM OLD.last_lease_expired_at AND
       EXISTS (
           SELECT 1 FROM platform.schema_job_steps AS step
           WHERE step.job_id = OLD.job_id
             AND step.ordinal = OLD.next_step_ordinal
             AND step.state IN ('pending', 'retryable', 'lease_expired')
             AND step.executor_kind = OLD.next_executor
       ) THEN
        RETURN NEW;
    END IF;

    IF OLD.state = 'leased' AND NEW.state = 'running' AND
       NEW.next_step_ordinal = OLD.next_step_ordinal AND
       NEW.next_executor = OLD.next_executor AND NEW.attempt = OLD.attempt AND
       NEW.available_at = OLD.available_at AND
       ROW(NEW.lease_id, NEW.lease_owner_id, NEW.leased_at,
           NEW.lease_heartbeat_at, NEW.lease_expires_at)
       IS NOT DISTINCT FROM
       ROW(OLD.lease_id, OLD.lease_owner_id, OLD.leased_at,
           OLD.lease_heartbeat_at, OLD.lease_expires_at) AND
       NEW.last_lease_expired_at IS NOT DISTINCT FROM OLD.last_lease_expired_at AND
       EXISTS (
           SELECT 1 FROM platform.schema_job_steps AS step
           WHERE step.job_id = OLD.job_id AND step.ordinal = OLD.next_step_ordinal
             AND step.state = 'running' AND step.executor_kind = OLD.next_executor
       ) THEN
        RETURN NEW;
    END IF;

    IF OLD.state IN ('leased', 'running') AND NEW.state = 'lease_expired' AND
       NEW.next_step_ordinal = OLD.next_step_ordinal AND
       NEW.next_executor = OLD.next_executor AND NEW.attempt = OLD.attempt AND
       NEW.available_at = OLD.available_at AND
       ROW(NEW.lease_id, NEW.lease_owner_id, NEW.leased_at,
           NEW.lease_heartbeat_at, NEW.lease_expires_at)
       IS NOT DISTINCT FROM
       ROW(OLD.lease_id, OLD.lease_owner_id, OLD.leased_at,
           OLD.lease_heartbeat_at, OLD.lease_expires_at) AND
       NEW.last_lease_expired_at = NEW.updated_at AND
       OLD.lease_expires_at <= NEW.updated_at AND EXISTS (
           SELECT 1 FROM platform.schema_job_steps AS step
           WHERE step.job_id = OLD.job_id AND step.ordinal = OLD.next_step_ordinal
             AND step.state = 'lease_expired'
       ) THEN
        RETURN NEW;
    END IF;

    IF OLD.state = 'lease_expired' AND
       ((OLD.next_executor = 'migrator' AND NEW.state = 'queued') OR
        (OLD.next_executor = 'forms_runtime' AND NEW.state = 'awaiting_runtime')) AND
       NEW.next_step_ordinal = OLD.next_step_ordinal AND
       NEW.next_executor = OLD.next_executor AND NEW.attempt = OLD.attempt AND
       NEW.available_at >= NEW.updated_at AND
       NEW.last_lease_expired_at = OLD.last_lease_expired_at THEN
        RETURN NEW;
    END IF;

    IF OLD.state = 'running' AND
       NEW.state IN ('queued', 'awaiting_runtime') AND
       NEW.next_step_ordinal = OLD.next_step_ordinal AND
       NEW.attempt = OLD.attempt AND NEW.available_at >= NEW.updated_at AND
       NEW.last_lease_expired_at IS NOT DISTINCT FROM OLD.last_lease_expired_at AND
       ((NEW.state = 'queued' AND NEW.next_executor = 'migrator') OR
        (NEW.state = 'awaiting_runtime' AND NEW.next_executor = 'forms_runtime')) AND
       EXISTS (
           SELECT 1 FROM platform.schema_job_steps AS step
           WHERE step.job_id = OLD.job_id AND step.ordinal = OLD.next_step_ordinal
             AND step.state = 'retryable' AND step.executor_kind = NEW.next_executor
       ) THEN
        RETURN NEW;
    END IF;

    IF OLD.state = 'running' AND
       NEW.state IN ('queued', 'awaiting_runtime', 'succeeded') AND
       NEW.next_step_ordinal = OLD.next_step_ordinal + 1 AND
       NEW.attempt = OLD.attempt AND
       NEW.last_lease_expired_at IS NOT DISTINCT FROM OLD.last_lease_expired_at AND
       EXISTS (
           SELECT 1 FROM platform.schema_job_steps AS completed
           WHERE completed.job_id = OLD.job_id
             AND completed.ordinal = OLD.next_step_ordinal
             AND completed.state IN ('succeeded', 'skipped')
       ) AND (
           (NEW.next_step_ordinal = NEW.step_count AND NEW.state = 'succeeded' AND
               NEW.next_executor = 'none' AND NEW.finished_at = NEW.updated_at) OR
           (NEW.next_step_ordinal < NEW.step_count AND NEW.available_at >= NEW.updated_at AND
               EXISTS (
                   SELECT 1 FROM platform.schema_job_steps AS next_step
                   WHERE next_step.job_id = OLD.job_id
                     AND next_step.ordinal = NEW.next_step_ordinal
                     AND next_step.state = 'pending'
                     AND next_step.executor_kind = NEW.next_executor
                     AND ((NEW.next_executor = 'migrator' AND NEW.state = 'queued') OR
                          (NEW.next_executor = 'forms_runtime' AND
                              NEW.state = 'awaiting_runtime'))
               ))
       ) THEN
        RETURN NEW;
    END IF;

    IF OLD.state = 'running' AND NEW.state IN ('failed', 'cancelled') AND
       NEW.next_step_ordinal = OLD.next_step_ordinal AND NEW.attempt = OLD.attempt AND
       NEW.finished_at = NEW.updated_at AND
       NEW.last_lease_expired_at IS NOT DISTINCT FROM OLD.last_lease_expired_at AND
       EXISTS (
           SELECT 1 FROM platform.schema_job_steps AS step
           WHERE step.job_id = OLD.job_id AND step.ordinal = OLD.next_step_ordinal AND
             ((NEW.state = 'failed' AND step.state IN ('failed', 'blocked')) OR
              (NEW.state = 'cancelled' AND step.state = 'cancelled'))
       ) THEN
        RETURN NEW;
    END IF;

    IF OLD.state IN ('queued', 'awaiting_runtime', 'lease_expired') AND
       NEW.state = 'cancelled' AND NEW.next_step_ordinal = OLD.next_step_ordinal AND
       NEW.attempt = OLD.attempt AND NEW.available_at = OLD.available_at AND
       NEW.finished_at = NEW.updated_at AND
       NEW.last_lease_expired_at IS NOT DISTINCT FROM OLD.last_lease_expired_at THEN
        RETURN NEW;
    END IF;

    RAISE EXCEPTION USING
        ERRCODE = '55000', MESSAGE = 'invalid schema job transition';
END;
$function$;
-- +goose StatementEnd

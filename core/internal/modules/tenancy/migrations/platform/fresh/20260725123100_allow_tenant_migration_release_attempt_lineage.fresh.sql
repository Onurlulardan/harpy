-- +goose Up
ALTER TABLE platform.tenancy_tenant_migration_jobs
DROP CONSTRAINT tenancy_tenant_migration_jobs_current_attempt_check;
ALTER TABLE platform.tenancy_tenant_migration_jobs
ADD CONSTRAINT tenancy_tenant_migration_jobs_current_attempt_check
CHECK (current_attempt BETWEEN 0 AND 32767);

ALTER TABLE platform.tenancy_tenant_migration_attempts
DROP CONSTRAINT tenancy_tenant_migration_attempts_attempt_check;
ALTER TABLE platform.tenancy_tenant_migration_attempts
ADD CONSTRAINT tenancy_tenant_migration_attempts_attempt_check
CHECK (attempt BETWEEN 1 AND 32767);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION platform.tenancy_guard_tenant_migration_job_v1()
RETURNS trigger
LANGUAGE plpgsql SET search_path TO pg_catalog
AS $function$
DECLARE
    v_release_changed boolean;
    v_rollforward boolean := false;
    v_now timestamptz := pg_catalog.clock_timestamp();
BEGIN
    IF TG_OP = 'DELETE' OR ROW(
        NEW.job_id, NEW.operation_id, NEW.origin, NEW.environment_id,
        NEW.deployment_mode, NEW.created_at
    ) IS DISTINCT FROM ROW(
        OLD.job_id, OLD.operation_id, OLD.origin, OLD.environment_id,
        OLD.deployment_mode, OLD.created_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration job identity is immutable';
    END IF;

    v_release_changed := ROW(
        NEW.release_id, NEW.runtime_generation, NEW.database_generation,
        NEW.migration_manifest_sha256
    ) IS DISTINCT FROM ROW(
        OLD.release_id, OLD.runtime_generation, OLD.database_generation,
        OLD.migration_manifest_sha256
    );
    IF v_release_changed THEN
        v_rollforward :=
            SESSION_USER = 'forgecrud_migrator'
            AND CURRENT_USER = 'forgecrud_owner'
            AND OLD.state IN ('pending', 'failed', 'running')
            AND NEW.state = 'running'
            AND NEW.current_attempt = OLD.current_attempt + 1
            AND NEW.error_code IS NULL AND NEW.completed_at IS NULL
            AND EXISTS (
                SELECT 1
                FROM platform.tenancy_tenant_migration_executor_readiness AS ready
                WHERE ready.release_id = NEW.release_id
                  AND ready.environment_id = NEW.environment_id
                  AND ready.deployment_mode = NEW.deployment_mode
                  AND ready.runtime_generation = NEW.runtime_generation
                  AND ready.database_generation = NEW.database_generation
                  AND ready.migration_manifest_sha256 =
                      NEW.migration_manifest_sha256
                  AND ready.state = 'active' AND ready.ready_until > v_now
            )
            AND (
                OLD.state <> 'running' OR EXISTS (
                    SELECT 1
                    FROM platform.tenancy_tenant_migration_attempts AS attempt
                    WHERE attempt.job_id = OLD.job_id
                      AND attempt.attempt = OLD.current_attempt
                      AND attempt.state = 'failed'
                      AND attempt.error_code = 'lease_expired'
                )
            )
            AND (
                SELECT pg_catalog.count(*)
                FROM platform.tenancy_tenant_migration_attempts AS attempt
                JOIN platform.tenancy_tenant_migration_executor_readiness AS ready
                  ON ready.profile_instance_id = attempt.profile_instance_id
                WHERE attempt.job_id = OLD.job_id
                  AND ready.release_id = NEW.release_id
                  AND ready.runtime_generation = NEW.runtime_generation
                  AND ready.database_generation = NEW.database_generation
                  AND ready.migration_manifest_sha256 =
                      NEW.migration_manifest_sha256
            ) < 8;
        IF NOT v_rollforward THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'tenant migration job release identity is immutable';
        END IF;
    END IF;

    IF NEW.updated_at <= OLD.updated_at OR NOT (
        OLD.state = 'pending' AND NEW.state IN ('running', 'fenced')
        OR OLD.state = 'running' AND NEW.state = 'running'
            AND NEW.current_attempt = OLD.current_attempt + 1
        OR OLD.state = 'running' AND NEW.state IN (
            'succeeded', 'failed', 'reconcile_required', 'cleanup_required'
        )
        OR OLD.state = 'failed' AND NEW.state IN ('running', 'fenced')
        OR OLD.state IN ('reconcile_required', 'cleanup_required')
            AND NEW.state = 'fenced'
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration job transition is invalid';
    END IF;
    IF NEW.state = 'running' AND NEW.current_attempt <> OLD.current_attempt + 1
       OR NEW.state <> 'running' AND
          NEW.current_attempt <> OLD.current_attempt THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration attempt transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

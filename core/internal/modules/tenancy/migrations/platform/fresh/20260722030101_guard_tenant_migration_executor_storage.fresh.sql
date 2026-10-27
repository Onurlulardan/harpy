-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.tenancy_guard_tenant_migration_readiness_v1()
RETURNS trigger
LANGUAGE plpgsql SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' OR ROW(
        NEW.profile_instance_id, NEW.release_id, NEW.environment_id,
        NEW.deployment_mode, NEW.runtime_generation, NEW.database_generation,
        NEW.migration_manifest_sha256, NEW.created_at
    ) IS DISTINCT FROM ROW(
        OLD.profile_instance_id, OLD.release_id, OLD.environment_id,
        OLD.deployment_mode, OLD.runtime_generation, OLD.database_generation,
        OLD.migration_manifest_sha256, OLD.created_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration readiness identity is immutable';
    END IF;
    IF NOT (
        OLD.state = 'active' AND NEW.state = 'active'
            AND NEW.heartbeat_at > OLD.heartbeat_at
            AND NEW.ready_until > OLD.ready_until
        OR OLD.state = 'active' AND NEW.state = 'retired'
            AND NEW.retired_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration readiness transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;

CREATE FUNCTION platform.tenancy_guard_tenant_migration_job_v1()
RETURNS trigger
LANGUAGE plpgsql SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' OR ROW(
        NEW.job_id, NEW.operation_id, NEW.origin, NEW.release_id,
        NEW.environment_id, NEW.deployment_mode, NEW.runtime_generation,
        NEW.database_generation, NEW.migration_manifest_sha256, NEW.created_at
    ) IS DISTINCT FROM ROW(
        OLD.job_id, OLD.operation_id, OLD.origin, OLD.release_id,
        OLD.environment_id, OLD.deployment_mode, OLD.runtime_generation,
        OLD.database_generation, OLD.migration_manifest_sha256, OLD.created_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration job identity is immutable';
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

CREATE FUNCTION platform.tenancy_guard_tenant_migration_attempt_v1()
RETURNS trigger
LANGUAGE plpgsql SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' OR ROW(
        NEW.job_id, NEW.attempt, NEW.migration_run_id,
        NEW.profile_instance_id, NEW.lease_id, NEW.leased_at
    ) IS DISTINCT FROM ROW(
        OLD.job_id, OLD.attempt, OLD.migration_run_id,
        OLD.profile_instance_id, OLD.lease_id, OLD.leased_at
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'tenant migration attempt identity is immutable';
    END IF;
    IF NOT (
        OLD.state = 'running' AND NEW.state = 'running'
            AND NEW.heartbeat_at > OLD.heartbeat_at
            AND NEW.lease_until > OLD.lease_until
        OR OLD.state = 'running' AND NEW.state IN (
            'succeeded', 'failed', 'reconcile_required', 'cleanup_required'
        ) AND NEW.finished_at IS NOT NULL
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '40001',
            MESSAGE = 'tenant migration attempt transition is invalid';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER tenancy_tenant_migration_readiness_guard
BEFORE UPDATE OR DELETE
ON platform.tenancy_tenant_migration_executor_readiness
FOR EACH ROW EXECUTE FUNCTION
    platform.tenancy_guard_tenant_migration_readiness_v1();

CREATE TRIGGER tenancy_tenant_migration_jobs_guard
BEFORE UPDATE OR DELETE ON platform.tenancy_tenant_migration_jobs
FOR EACH ROW EXECUTE FUNCTION platform.tenancy_guard_tenant_migration_job_v1();

CREATE TRIGGER tenancy_tenant_migration_attempts_guard
BEFORE UPDATE OR DELETE ON platform.tenancy_tenant_migration_attempts
FOR EACH ROW EXECUTE FUNCTION
    platform.tenancy_guard_tenant_migration_attempt_v1();

CREATE TRIGGER tenancy_tenant_migration_receipts_guard
BEFORE UPDATE OR DELETE ON platform.tenancy_tenant_migration_receipts
FOR EACH ROW EXECUTE FUNCTION
    platform.tenancy_reject_onboarding_append_mutation_v1();

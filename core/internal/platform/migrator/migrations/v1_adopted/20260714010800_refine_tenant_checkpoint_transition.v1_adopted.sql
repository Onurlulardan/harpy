-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_checkpoint_transition_v2()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.state <> 'pending' OR NEW.attempt <> 1 OR
           NEW.started_at IS NOT NULL OR NEW.finished_at IS NOT NULL OR
           NEW.error_code IS NOT NULL THEN
            RAISE EXCEPTION USING
                ERRCODE = '55000', MESSAGE = 'invalid initial checkpoint';
        END IF;
        NEW.updated_at := pg_catalog.clock_timestamp();
        RETURN NEW;
    END IF;
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'checkpoint deletion is forbidden';
    END IF;
    IF ROW(NEW.tenant_id, NEW.owner, NEW.logical_version, NEW.database_origin,
           NEW.selected_path, NEW.selected_sha256)
       IS DISTINCT FROM
       ROW(OLD.tenant_id, OLD.owner, OLD.logical_version, OLD.database_origin,
           OLD.selected_path, OLD.selected_sha256) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'checkpoint identity is immutable';
    END IF;
    IF NOT (
        (OLD.state = 'pending' AND NEW.state = 'running' AND
            NEW.attempt = OLD.attempt AND NEW.run_id = OLD.run_id) OR
        (OLD.state = 'failed' AND NEW.state = 'running' AND
            NEW.attempt = OLD.attempt + 1 AND NEW.run_id <> OLD.run_id) OR
        (OLD.state = 'running' AND NEW.state IN ('succeeded', 'failed', 'fenced') AND
            NEW.attempt = OLD.attempt AND NEW.run_id = OLD.run_id)
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'invalid checkpoint transition';
    END IF;
    IF NEW.state = 'running' THEN
        NEW.started_at := pg_catalog.clock_timestamp();
        NEW.finished_at := NULL;
        NEW.error_code := NULL;
    ELSE
        IF NOT EXISTS (
            SELECT 1
            FROM platform.tenant_migration_checkpoint_attempts AS attempt
            WHERE attempt.tenant_id = NEW.tenant_id
              AND attempt.owner = NEW.owner
              AND attempt.logical_version = NEW.logical_version
              AND attempt.attempt = NEW.attempt
              AND attempt.database_origin = NEW.database_origin
              AND attempt.selected_path = NEW.selected_path
              AND attempt.selected_sha256 = NEW.selected_sha256
              AND attempt.run_id = NEW.run_id
              AND attempt.state = NEW.state
        ) THEN
            RAISE EXCEPTION USING
                ERRCODE = '55000', MESSAGE = 'checkpoint receipt is missing';
        END IF;
        NEW.started_at := (
            SELECT attempt.started_at
            FROM platform.tenant_migration_checkpoint_attempts AS attempt
            WHERE attempt.tenant_id = NEW.tenant_id AND attempt.owner = NEW.owner
              AND attempt.logical_version = NEW.logical_version
              AND attempt.attempt = NEW.attempt
        );
        NEW.finished_at := (
            SELECT attempt.finished_at
            FROM platform.tenant_migration_checkpoint_attempts AS attempt
            WHERE attempt.tenant_id = NEW.tenant_id AND attempt.owner = NEW.owner
              AND attempt.logical_version = NEW.logical_version
              AND attempt.attempt = NEW.attempt
        );
        NEW.error_code := (
            SELECT attempt.error_code
            FROM platform.tenant_migration_checkpoint_attempts AS attempt
            WHERE attempt.tenant_id = NEW.tenant_id AND attempt.owner = NEW.owner
              AND attempt.logical_version = NEW.logical_version
              AND attempt.attempt = NEW.attempt
        );
    END IF;
    NEW.updated_at := pg_catalog.clock_timestamp();
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

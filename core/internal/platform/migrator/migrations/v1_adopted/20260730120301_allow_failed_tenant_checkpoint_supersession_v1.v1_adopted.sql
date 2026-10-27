-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_checkpoint_transition_v3()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
DECLARE
    v_identity_changed boolean;
    v_registered_supersession boolean;
BEGIN
    IF TG_OP = 'INSERT' THEN
        IF NEW.state <> 'pending' OR NEW.attempt <> 1 OR
           NEW.started_at IS NOT NULL OR NEW.finished_at IS NOT NULL OR
           NEW.error_code IS NOT NULL THEN
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'invalid initial checkpoint';
        END IF;
        NEW.updated_at := pg_catalog.clock_timestamp();
        RETURN NEW;
    END IF;
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'checkpoint deletion is forbidden';
    END IF;
    v_identity_changed :=
        ROW(NEW.tenant_id,NEW.owner,NEW.logical_version,NEW.database_origin,
            NEW.selected_path,NEW.selected_sha256)
        IS DISTINCT FROM
        ROW(OLD.tenant_id,OLD.owner,OLD.logical_version,OLD.database_origin,
            OLD.selected_path,OLD.selected_sha256);
    v_registered_supersession := v_identity_changed
        AND OLD.state = 'failed' AND NEW.state = 'running'
        AND NEW.attempt = OLD.attempt + 1 AND NEW.run_id <> OLD.run_id
        AND NEW.tenant_id = OLD.tenant_id AND NEW.owner = OLD.owner
        AND NEW.logical_version = OLD.logical_version
        AND EXISTS (
            SELECT 1
            FROM platform.tenant_migration_supersessions AS supersession
            WHERE supersession.owner = NEW.owner
              AND supersession.scope = 'tenant'
              AND supersession.logical_version = NEW.logical_version
              AND supersession.database_origin = NEW.database_origin
              AND supersession.database_origin = OLD.database_origin
              AND supersession.superseded_path = OLD.selected_path
              AND supersession.superseded_sha256 = OLD.selected_sha256
              AND supersession.current_path = NEW.selected_path
              AND supersession.current_sha256 = NEW.selected_sha256
        );
    IF v_identity_changed AND NOT v_registered_supersession THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'checkpoint identity is immutable';
    END IF;
    IF NOT (
        (OLD.state = 'pending' AND NEW.state = 'running' AND
            NEW.attempt = OLD.attempt AND NEW.run_id = OLD.run_id) OR
        (OLD.state = 'failed' AND NEW.state = 'running' AND
            NEW.attempt = OLD.attempt + 1 AND NEW.run_id <> OLD.run_id) OR
        (OLD.state = 'running' AND NEW.state IN ('succeeded','failed','fenced') AND
            NEW.attempt = OLD.attempt AND NEW.run_id = OLD.run_id)
    ) THEN
        RAISE EXCEPTION USING ERRCODE = '55000',
            MESSAGE = 'invalid checkpoint transition';
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
            RAISE EXCEPTION USING ERRCODE = '55000',
                MESSAGE = 'checkpoint receipt is missing';
        END IF;
        SELECT attempt.started_at,attempt.finished_at,attempt.error_code
        INTO NEW.started_at,NEW.finished_at,NEW.error_code
        FROM platform.tenant_migration_checkpoint_attempts AS attempt
        WHERE attempt.tenant_id = NEW.tenant_id
          AND attempt.owner = NEW.owner
          AND attempt.logical_version = NEW.logical_version
          AND attempt.attempt = NEW.attempt;
    END IF;
    NEW.updated_at := pg_catalog.clock_timestamp();
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

ALTER TABLE platform.tenant_migration_checkpoints
    DISABLE TRIGGER tenant_migration_checkpoints_validate_transition_v2;

CREATE TRIGGER tenant_migration_checkpoints_validate_transition_v3
BEFORE INSERT OR UPDATE OR DELETE ON platform.tenant_migration_checkpoints
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_validate_checkpoint_transition_v3();

ALTER TABLE platform.tenant_migration_checkpoints
    ENABLE ALWAYS TRIGGER tenant_migration_checkpoints_validate_transition_v3;

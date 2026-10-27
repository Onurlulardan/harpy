-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_tenant_checkpoint_attempt()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP <> 'INSERT' THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'tenant checkpoint attempt is append only';
    END IF;
    IF NOT EXISTS (
        SELECT 1
        FROM platform.tenant_migration_checkpoints AS checkpoint
        WHERE checkpoint.tenant_id = NEW.tenant_id
          AND checkpoint.owner = NEW.owner
          AND checkpoint.logical_version = NEW.logical_version
          AND checkpoint.database_origin = NEW.database_origin
          AND checkpoint.selected_path = NEW.selected_path
          AND checkpoint.selected_sha256 = NEW.selected_sha256
          AND checkpoint.run_id = NEW.run_id
          AND checkpoint.state = 'running'
          AND checkpoint.attempt = NEW.attempt
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'tenant checkpoint attempt is invalid';
    END IF;
    NEW.started_at := (
        SELECT checkpoint.started_at
        FROM platform.tenant_migration_checkpoints AS checkpoint
        WHERE checkpoint.tenant_id = NEW.tenant_id
          AND checkpoint.owner = NEW.owner
          AND checkpoint.logical_version = NEW.logical_version
    );
    NEW.finished_at := pg_catalog.clock_timestamp();
    NEW.recorded_at := NEW.finished_at;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

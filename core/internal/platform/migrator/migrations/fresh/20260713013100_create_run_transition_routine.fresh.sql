-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_run_transition()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'migration run deletion is forbidden';
    END IF;
    IF ROW(NEW.run_id, NEW.release_id, NEW.environment_id, NEW.database_origin,
           NEW.command, NEW.migration_manifest_sha256, NEW.extension_manifest_sha256,
           NEW.roles_manifest_sha256, NEW.privileges_manifest_sha256,
           NEW.ownership_manifest_sha256, NEW.created_at)
       IS DISTINCT FROM
       ROW(OLD.run_id, OLD.release_id, OLD.environment_id, OLD.database_origin,
           OLD.command, OLD.migration_manifest_sha256, OLD.extension_manifest_sha256,
           OLD.roles_manifest_sha256, OLD.privileges_manifest_sha256,
           OLD.ownership_manifest_sha256, OLD.created_at) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'migration run identity is immutable';
    END IF;
    IF NOT ((OLD.state = 'planned' AND NEW.state IN ('running', 'fenced'))
        OR (OLD.state = 'running' AND NEW.state IN ('succeeded', 'failed', 'fenced'))) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'invalid migration run transition';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

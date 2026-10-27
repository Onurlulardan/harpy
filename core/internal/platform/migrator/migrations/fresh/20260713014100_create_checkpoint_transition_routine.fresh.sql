-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_checkpoint_transition()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'checkpoint deletion is forbidden';
    END IF;
    IF ROW(NEW.tenant_id, NEW.owner, NEW.logical_version, NEW.database_origin,
           NEW.selected_path, NEW.selected_sha256, NEW.run_id)
       IS DISTINCT FROM
       ROW(OLD.tenant_id, OLD.owner, OLD.logical_version, OLD.database_origin,
           OLD.selected_path, OLD.selected_sha256, OLD.run_id) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'checkpoint identity is immutable';
    END IF;
    IF NOT ((OLD.state = 'pending' AND NEW.state = 'running' AND NEW.attempt = OLD.attempt)
        OR (OLD.state = 'running' AND NEW.state IN ('succeeded', 'failed', 'fenced')
            AND NEW.attempt = OLD.attempt)
        OR (OLD.state = 'failed' AND NEW.state = 'running' AND NEW.attempt = OLD.attempt + 1)
        OR (OLD.state = 'failed' AND NEW.state = 'fenced' AND NEW.attempt = OLD.attempt)) THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'invalid checkpoint transition';
    END IF;
    IF NEW.updated_at <= OLD.updated_at THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'checkpoint time must advance';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_contract_transition()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'schema contract deletion is forbidden';
    END IF;
    IF NEW.component IS DISTINCT FROM OLD.component THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'schema component is immutable';
    END IF;
    IF NEW.revision <> OLD.revision + 1
        OR NEW.database_generation < OLD.database_generation
        OR NEW.minimum_runtime_generation < OLD.minimum_runtime_generation
        OR NEW.verified_at <= OLD.verified_at
        OR NEW.verified_run_id = OLD.verified_run_id THEN
        RAISE EXCEPTION USING ERRCODE = '55000', MESSAGE = 'invalid schema contract transition';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

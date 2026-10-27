-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_migrator_validate_schema_job_revision()
RETURNS trigger
LANGUAGE plpgsql
SET search_path TO pg_catalog
AS $function$
BEGIN
    IF NOT pg_catalog.pg_try_advisory_xact_lock(pg_catalog.hashtextextended(
        'forgecrud:v2:schema-revision:' || NEW.tenant_id::text || ':' ||
        NEW.owner || ':' || NEW.resource_kind || ':' || NEW.resource_id::text,
        0
    )) THEN
        RAISE EXCEPTION USING
            ERRCODE = '40001', MESSAGE = 'schema resource revision serialization conflict';
    END IF;
    IF EXISTS (
        SELECT 1
        FROM platform.schema_jobs AS previous
        WHERE previous.tenant_id = NEW.tenant_id
          AND previous.owner = NEW.owner
          AND previous.resource_kind = NEW.resource_kind
          AND previous.resource_id = NEW.resource_id
          AND previous.resource_revision >= NEW.resource_revision
    ) THEN
        RAISE EXCEPTION USING
            ERRCODE = '55000', MESSAGE = 'schema resource revision is not monotonic';
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

CREATE TRIGGER schema_jobs_validate_revision
BEFORE INSERT ON platform.schema_jobs
FOR EACH ROW
EXECUTE FUNCTION platform.platform_migrator_validate_schema_job_revision();

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_guard_job_transition_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP IS DISTINCT FROM 'UPDATE' OR
       TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'jobs' THEN
        RAISE EXCEPTION USING ERRCODE = 'P5610',
            MESSAGE = 'job transition invocation is invalid';
    END IF;
    IF platform.platform_async_validate_job_transition_v1(
        OLD, NEW, CURRENT_USER, SESSION_USER
    ) IS NULL THEN NULL; END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_job_transition_v1() FROM PUBLIC;

CREATE TRIGGER jobs_transition_guard
BEFORE UPDATE ON platform.jobs
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_job_transition_v1();
ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_transition_guard;

-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_guard_job_attempt_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_TABLE_SCHEMA IS DISTINCT FROM 'platform' OR
       TG_TABLE_NAME IS DISTINCT FROM 'job_attempts' OR
       NOT (TG_OP = ANY (ARRAY['INSERT','UPDATE']::text[])) THEN
        RAISE EXCEPTION USING ERRCODE = 'P5620',
            MESSAGE = 'job attempt invocation is invalid';
    END IF;
    IF TG_OP = 'INSERT' THEN
        IF platform.platform_async_validate_job_attempt_v1(
            NULL::platform.job_attempts, NEW, CURRENT_USER, SESSION_USER
        ) IS NULL THEN NULL; END IF;
    ELSE
        IF platform.platform_async_validate_job_attempt_v1(
            OLD, NEW, CURRENT_USER, SESSION_USER
        ) IS NULL THEN NULL; END IF;
    END IF;
    RETURN NEW;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_job_attempt_v1() FROM PUBLIC;

CREATE TRIGGER job_attempts_transition_guard
BEFORE INSERT OR UPDATE ON platform.job_attempts
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_job_attempt_v1();
ALTER TABLE platform.job_attempts ENABLE ALWAYS TRIGGER job_attempts_transition_guard;

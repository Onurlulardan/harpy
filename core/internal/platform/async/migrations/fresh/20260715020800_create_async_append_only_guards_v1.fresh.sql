-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_guard_append_only_v1()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog
AS $function$
BEGIN
    IF TG_OP OPERATOR(pg_catalog.=) 'DELETE' AND
       CURRENT_USER OPERATOR(pg_catalog.=) 'forgecrud_owner' THEN
        RETURN OLD;
    END IF;
    RAISE EXCEPTION USING
        ERRCODE = '55000',
        MESSAGE = 'platform async record is append-only';
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION platform.platform_async_guard_append_only_v1() FROM PUBLIC;

CREATE TRIGGER jobs_delete_guard
BEFORE DELETE ON platform.jobs
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();

CREATE TRIGGER job_attempts_delete_guard
BEFORE DELETE ON platform.job_attempts
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();

CREATE TRIGGER outbox_delete_guard
BEFORE DELETE ON platform.outbox
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();

CREATE TRIGGER inbox_append_only_guard
BEFORE UPDATE OR DELETE ON platform.inbox
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();

CREATE TRIGGER idempotency_receipts_append_only_guard
BEFORE UPDATE OR DELETE ON platform.idempotency_receipts
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();

CREATE TRIGGER dead_letters_delete_guard
BEFORE DELETE ON platform.dead_letters
FOR EACH ROW
EXECUTE FUNCTION platform.platform_async_guard_append_only_v1();

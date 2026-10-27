-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION platform.platform_async_run_retention_cycle_v1()
RETURNS TABLE (
    lock_acquired boolean,
    jobs_redacted integer,
    outbox_redacted integer,
    dead_letters_redacted integer,
    inbox_purged integer,
    idempotency_purged integer,
    jobs_purged integer,
    dead_letters_purged integer
)
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime' OR
       CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE = '42501',
            MESSAGE = 'async retention cycle requires the exact runtime authority';
    END IF;

    SELECT pg_catalog.pg_try_advisory_xact_lock(
        pg_catalog.hashtextextended(
            'harpy.platform_async.retention-cycle.v1', 0
        )
    ) INTO lock_acquired;
    jobs_redacted := 0;
    outbox_redacted := 0;
    dead_letters_redacted := 0;
    inbox_purged := 0;
    idempotency_purged := 0;
    jobs_purged := 0;
    dead_letters_purged := 0;

    IF NOT lock_acquired THEN
        RETURN NEXT;
        RETURN;
    END IF;

    jobs_redacted := platform.platform_async_redact_job_payloads_v1(500);
    outbox_redacted := platform.platform_async_redact_outbox_payloads_v2(500);
    dead_letters_redacted :=
        platform.platform_async_redact_dead_letter_payloads_v2(500);
    inbox_purged := platform.platform_async_purge_inbox_v1(500);
    idempotency_purged :=
        platform.platform_async_purge_idempotency_receipts_v1(500);
    jobs_purged := platform.platform_async_purge_jobs_v1(500);
    dead_letters_purged :=
        platform.platform_async_purge_dead_letters_v2(500);
    RETURN NEXT;
END;
$function$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION
    platform.platform_async_run_retention_cycle_v1() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION
    platform.platform_async_run_retention_cycle_v1() TO harpy_runtime;

REVOKE EXECUTE ON FUNCTION
    platform.platform_async_redact_job_payloads_v1(integer)
FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION
    platform.platform_async_redact_outbox_payloads_v2(integer)
FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION
    platform.platform_async_redact_dead_letter_payloads_v2(integer)
FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION
    platform.platform_async_purge_inbox_v1(integer)
FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION
    platform.platform_async_purge_idempotency_receipts_v1(integer)
FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION
    platform.platform_async_purge_jobs_v1(integer)
FROM harpy_runtime;
REVOKE EXECUTE ON FUNCTION
    platform.platform_async_purge_dead_letters_v2(integer)
FROM harpy_runtime;

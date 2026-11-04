-- +goose Up
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
LANGUAGE sql
VOLATILE
SECURITY INVOKER
SET search_path TO pg_catalog
AS $$
    SELECT
        false,
        0::integer,
        0::integer,
        0::integer,
        0::integer,
        0::integer,
        0::integer,
        0::integer;
$$;

REVOKE ALL ON FUNCTION platform.platform_async_run_retention_cycle_v1() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION platform.platform_async_run_retention_cycle_v1() TO harpy_api, harpy_worker;

-- +goose Down
DROP FUNCTION platform.platform_async_run_retention_cycle_v1();

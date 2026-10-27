SELECT
    lock_acquired,
    jobs_redacted,
    outbox_redacted,
    dead_letters_redacted,
    inbox_purged,
    idempotency_purged,
    jobs_purged,
    dead_letters_purged
FROM platform.platform_async_run_retention_cycle_v1()

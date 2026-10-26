WITH observation_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
), outbox_stats AS MATERIALIZED (
    SELECT
        MIN(event.created_at) FILTER (
            WHERE event.delivery_state <> 'delivered'
        ) AS oldest_undelivered_at,
        COUNT(*) FILTER (WHERE event.parked_at IS NOT NULL)::bigint AS parked
    FROM platform.outbox AS event
), job_stats AS MATERIALIZED (
    SELECT
        MIN(job.run_at) FILTER (
            WHERE job.state = 'queued'
              AND job.parked_at IS NULL
              AND job.run_at <= observation_clock.observed_at
        ) AS oldest_due_at,
        COUNT(*) FILTER (WHERE job.parked_at IS NOT NULL)::bigint AS parked
    FROM platform.jobs AS job
    CROSS JOIN observation_clock
)
SELECT
    COALESCE(GREATEST(0,
        pg_catalog.floor(EXTRACT(EPOCH FROM (
            observation_clock.observed_at - outbox_stats.oldest_undelivered_at
        )))
    )::bigint, 0::bigint) AS oldest_undelivered_outbox_age_seconds,
    CASE WHEN outbox_stats.oldest_undelivered_at <=
        observation_clock.observed_at - ($1::integer * interval '1 second')
        THEN 1::bigint ELSE 0::bigint
    END AS undelivered_outbox_alarm,
    COALESCE(GREATEST(0,
        pg_catalog.floor(EXTRACT(EPOCH FROM (
            observation_clock.observed_at - job_stats.oldest_due_at
        )))
    )::bigint, 0::bigint) AS oldest_due_job_lag_seconds,
    job_stats.parked AS parked_jobs,
    outbox_stats.parked AS parked_outbox
FROM observation_clock
CROSS JOIN outbox_stats
CROSS JOIN job_stats;

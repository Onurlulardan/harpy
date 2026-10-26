WITH cancel_clock AS MATERIALIZED (
    SELECT pg_catalog.clock_timestamp() AS observed_at
), changed AS (
    UPDATE platform.jobs AS job
    SET state = CASE WHEN job.state = 'queued' THEN 'cancelled' ELSE job.state END,
        cancel_requested_at = CASE
            WHEN job.state = 'running' THEN GREATEST(
                cancel_clock.observed_at,
                job.updated_at + interval '1 microsecond'
            )
            ELSE job.cancel_requested_at
        END,
        last_error_class = CASE
            WHEN job.state = 'queued' THEN 'cancelled' ELSE job.last_error_class
        END,
        last_error_code = CASE
            WHEN job.state = 'queued' THEN $8 ELSE job.last_error_code
        END,
        last_error_sha256 = CASE
            WHEN job.state = 'queued' THEN $9 ELSE job.last_error_sha256
        END,
        retain_until = CASE
            WHEN job.state = 'queued' THEN GREATEST(
                cancel_clock.observed_at,
                job.updated_at + interval '1 microsecond'
            ) + interval '90 days'
            ELSE job.retain_until
        END,
        updated_at = GREATEST(
            cancel_clock.observed_at,
            job.updated_at + interval '1 microsecond'
        ),
        completed_at = CASE
            WHEN job.state = 'queued' THEN GREATEST(
                cancel_clock.observed_at,
                job.updated_at + interval '1 microsecond'
            )
            ELSE job.completed_at
        END
    FROM cancel_clock
    WHERE job.scope_kind = $1
      AND job.scope_id = $2
      AND job.authority_group_id = $3
      AND job.authority_group_generation = $4
      AND job.authority_domain = $5
      AND job.cutover_generation = $6
      AND job.id = $7
      AND (job.state = 'queued' OR
           (job.state = 'running' AND job.cancel_requested_at IS NULL))
    RETURNING
        job.state,
        COALESCE(job.cancel_requested_at, job.completed_at) AS requested_at
)
SELECT state, requested_at
FROM changed
UNION ALL
SELECT job.state, job.cancel_requested_at
FROM platform.jobs AS job
WHERE NOT EXISTS (SELECT 1 FROM changed)
  AND job.scope_kind = $1
  AND job.scope_id = $2
  AND job.authority_group_id = $3
  AND job.authority_group_generation = $4
  AND job.authority_domain = $5
  AND job.cutover_generation = $6
  AND job.id = $7
  AND job.state = 'running'
  AND job.cancel_requested_at IS NOT NULL
LIMIT 1;

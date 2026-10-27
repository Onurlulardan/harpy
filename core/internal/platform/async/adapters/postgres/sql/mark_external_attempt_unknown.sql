WITH completion_clock AS MATERIALIZED (
    SELECT GREATEST(
        pg_catalog.clock_timestamp(),
        job.updated_at + interval '1 microsecond'
    ) AS completed_at
    FROM platform.jobs AS job
    WHERE job.id = $2
), completed AS (
    UPDATE platform.job_attempts AS attempt
    SET finished_at = completion_clock.completed_at,
        outcome = 'dead',
        error_class = NULL,
        error_code = $6,
        redacted_error_sha256 = NULL
    FROM completion_clock
    WHERE attempt.attempt_id = $1
      AND attempt.job_id = $2
      AND attempt.lease_owner_id = $3
      AND attempt.lease_generation = $4
      AND attempt.finished_at IS NULL
      AND attempt.effect_started_at = $5
      AND attempt.effect_id = $7
      AND attempt.effect_request_sha256 = $8
      AND attempt.effect_receipt_id IS NULL
    RETURNING attempt.finished_at
)
SELECT finished_at
FROM completed;

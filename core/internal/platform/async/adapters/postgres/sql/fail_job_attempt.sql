WITH failure_clock AS MATERIALIZED (
    SELECT GREATEST(
        pg_catalog.clock_timestamp(),
        job.updated_at + interval '1 microsecond'
    ) AS occurred_at
    FROM platform.jobs AS job
    WHERE job.id = $2
), failed AS (
    UPDATE platform.job_attempts AS attempt
    SET finished_at = failure_clock.occurred_at,
        outcome = $5,
        error_class = $6,
        error_code = $7,
        redacted_error_sha256 = $8,
        effect_receipt_id = $9
    FROM failure_clock
    WHERE attempt.attempt_id = $1
      AND attempt.job_id = $2
      AND attempt.lease_owner_id = $3
      AND attempt.lease_generation = $4
      AND attempt.finished_at IS NULL
      AND attempt.effect_started_at IS NOT DISTINCT FROM $10::timestamptz
    RETURNING attempt.finished_at
)
SELECT finished_at
FROM failed;

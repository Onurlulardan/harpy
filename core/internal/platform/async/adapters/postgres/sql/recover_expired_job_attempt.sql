UPDATE platform.job_attempts AS attempt
SET finished_at = $5,
    outcome = $6,
    error_class = $7,
    error_code = $8,
    redacted_error_sha256 = $9,
    effect_receipt_id = $10
WHERE attempt.attempt_id = $1
  AND attempt.job_id = $2
  AND attempt.lease_owner_id = $3
  AND attempt.lease_generation = $4
  AND attempt.finished_at IS NULL;

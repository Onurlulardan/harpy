UPDATE platform.job_attempts AS attempt
SET effect_started_at = $5,
    effect_id = $6,
    effect_request_sha256 = $7
WHERE attempt.attempt_id = $1
  AND attempt.job_id = $2
  AND attempt.lease_owner_id = $3
  AND attempt.lease_generation = $4
  AND attempt.finished_at IS NULL
  AND attempt.effect_started_at IS NULL
  AND attempt.effect_id IS NULL
  AND attempt.effect_request_sha256 IS NULL
RETURNING attempt.effect_started_at;

UPDATE notifications_retention_checkpoints
SET revision = $1, last_job_id = $2::uuid, last_attempt_id = $3::uuid,
    last_completed_at = $4
WHERE singleton AND revision = $5
RETURNING revision;

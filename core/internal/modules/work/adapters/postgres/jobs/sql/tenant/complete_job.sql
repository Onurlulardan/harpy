UPDATE work_job_receipts
SET state = 'SUCCEEDED', outcome_sha256 = $5, completed_at = $6
WHERE job_id = $1 AND job_kind = $2 AND payload_sha256 = $3
  AND lease_generation = $4 AND state = 'RUNNING'

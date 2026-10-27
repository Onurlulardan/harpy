INSERT INTO work_job_receipts (
    job_id, job_kind, payload_sha256, lease_generation, state,
    outcome_sha256, started_at, completed_at
) VALUES ($1,$2,$3,$4,'RUNNING',NULL,$5,NULL)
ON CONFLICT (job_id) DO NOTHING

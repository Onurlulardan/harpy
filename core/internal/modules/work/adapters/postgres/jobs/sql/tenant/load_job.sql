SELECT job_kind, payload_sha256, lease_generation, state
FROM work_job_receipts
WHERE job_id = $1

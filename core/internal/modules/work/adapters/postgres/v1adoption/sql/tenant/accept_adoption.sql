UPDATE work_adoption_runs
SET state = 'ACCEPTED', acceptance_sha256 = $2,
    target_count = $3, target_sha256 = $4, updated_at = transaction_timestamp()
WHERE run_id = $1 AND state = 'RECONCILING'

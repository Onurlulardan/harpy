UPDATE work_adoption_runs
SET state = $2, applied_lsn = CASE WHEN $2 = 'RECONCILING' THEN pg_current_wal_lsn() ELSE applied_lsn END,
    updated_at = transaction_timestamp()
WHERE run_id = $1 AND state = ANY($3::text[])

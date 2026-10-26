INSERT INTO work_adoption_objects (
    run_id, source_kind, source_id, target_kind, target_id,
    source_sha256, target_sha256, applied_lsn, state, issue_code, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, pg_current_wal_lsn(), 'COPIED', NULL, transaction_timestamp())
ON CONFLICT (run_id, source_kind, source_id, target_kind, target_id) DO UPDATE
SET applied_lsn = pg_current_wal_lsn(), updated_at = transaction_timestamp()
WHERE work_adoption_objects.source_sha256 = EXCLUDED.source_sha256
  AND work_adoption_objects.target_sha256 = EXCLUDED.target_sha256
  AND work_adoption_objects.state IN ('COPIED','MATCHED')

WITH saved AS (
    INSERT INTO notifications_adoption_checkpoints (
        run_id, source_table, last_source_id, adopted_count, source_checksum,
        target_checksum, updated_at
    ) VALUES (
        $1::uuid, 'notifications', $2::uuid, $3, $4::bytea, $5::bytea,
        pg_catalog.transaction_timestamp()
    )
    ON CONFLICT (run_id, source_table) DO UPDATE
    SET last_source_id = EXCLUDED.last_source_id,
        adopted_count = EXCLUDED.adopted_count,
        source_checksum = EXCLUDED.source_checksum,
        target_checksum = EXCLUDED.target_checksum,
        updated_at = EXCLUDED.updated_at
    RETURNING run_id, source_checksum, target_checksum
)
UPDATE notifications_adoption_runs run
SET lifecycle_state = 'verified', completed_at = pg_catalog.transaction_timestamp(),
    row_version = row_version + 1
FROM saved
WHERE run.id = saved.run_id AND run.lifecycle_state IN ('planned','applying')
  AND saved.source_checksum = saved.target_checksum
  AND NOT EXISTS (
      SELECT 1 FROM notifications_adoption_issues issue
      WHERE issue.run_id = run.id AND issue.resolved_at IS NULL
  )
RETURNING run.id;

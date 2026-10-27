UPDATE notifications_writer_authority
SET lifecycle_state = 'quiescing', row_version = row_version + 1,
    updated_at = pg_catalog.transaction_timestamp()
WHERE singleton AND writer = 'v1' AND lifecycle_state = 'writable'
  AND generation = $1 AND row_version = $2
  AND EXISTS (
      SELECT 1 FROM notifications_adoption_runs
      WHERE lifecycle_state = 'verified'
  )
RETURNING writer, lifecycle_state, generation, row_version;

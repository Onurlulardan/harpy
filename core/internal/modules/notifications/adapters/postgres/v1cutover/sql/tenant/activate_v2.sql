UPDATE notifications_writer_authority
SET writer = 'v2', lifecycle_state = 'writable', generation = generation + 1,
    row_version = row_version + 1,
    updated_at = pg_catalog.transaction_timestamp()
WHERE singleton AND writer = 'v1' AND lifecycle_state = 'fenced'
  AND generation = $1 AND row_version = $2
RETURNING writer, lifecycle_state, generation, row_version;

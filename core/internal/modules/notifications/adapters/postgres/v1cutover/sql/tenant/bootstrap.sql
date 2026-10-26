UPDATE notifications_writer_authority
SET writer = CASE WHEN $1 = 'fresh' THEN 'v2' ELSE 'v1' END,
    lifecycle_state = 'writable', generation = 1,
    row_version = row_version + 1,
    updated_at = pg_catalog.transaction_timestamp()
WHERE singleton AND writer = 'unassigned' AND lifecycle_state = 'bootstrap'
  AND generation = 0 AND $1 IN ('fresh','v1_adopted')
RETURNING writer, lifecycle_state, generation, row_version;

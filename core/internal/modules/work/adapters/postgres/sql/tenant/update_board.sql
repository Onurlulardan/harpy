UPDATE work_boards
SET name = $2,
    description = $3,
    lifecycle_state = $4,
    row_version = $5,
    updated_at = transaction_timestamp(),
    archived_at = CASE WHEN $4 = 'ARCHIVED' THEN transaction_timestamp() ELSE NULL END
WHERE id = $1 AND row_version = $6

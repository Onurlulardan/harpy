UPDATE work_labels
SET name = $3, color = $4, row_version = $5, updated_at = transaction_timestamp()
WHERE id = $1 AND board_id = $2 AND row_version = $6 AND deleted_at IS NULL

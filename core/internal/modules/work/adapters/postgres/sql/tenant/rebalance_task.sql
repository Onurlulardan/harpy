UPDATE work_tasks
SET position = $3, row_version = row_version + 1, updated_at = transaction_timestamp()
WHERE id = $1 AND column_id = $2 AND deleted_at IS NULL

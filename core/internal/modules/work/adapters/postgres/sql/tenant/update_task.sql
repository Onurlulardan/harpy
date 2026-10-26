UPDATE work_tasks
SET column_id = $2, status_category = $3, priority = $4, title = $5, description = $6,
    position = $7, row_version = $8, starts_at = $9, due_at = $10, deleted_at = $11,
    completed_at = CASE WHEN $3 = 'DONE' THEN COALESCE(completed_at, transaction_timestamp()) ELSE NULL END,
    cancelled_at = CASE WHEN $3 = 'CANCELLED' THEN COALESCE(cancelled_at, transaction_timestamp()) ELSE NULL END,
    updated_at = transaction_timestamp()
WHERE id = $1 AND row_version = $12

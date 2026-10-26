UPDATE work_checklist_items
SET item_text = $3, completed = $4, position = $5,
    row_version = $6, updated_at = transaction_timestamp()
WHERE id = $1 AND task_id = $2 AND row_version = $7 AND deleted_at IS NULL

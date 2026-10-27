INSERT INTO work_checklist_items (
    id, task_id, item_text, completed, position, row_version, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, transaction_timestamp(), transaction_timestamp())

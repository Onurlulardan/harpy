INSERT INTO work_tasks (
    id, board_id, column_id, creator_user_id, task_kind, status_category, priority, title, description,
    position, row_version, starts_at, due_at, completed_at, cancelled_at, deleted_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
    CASE WHEN $6 = 'DONE' THEN transaction_timestamp() ELSE NULL END,
    CASE WHEN $6 = 'CANCELLED' THEN transaction_timestamp() ELSE NULL END,
    NULL, transaction_timestamp(), transaction_timestamp()
)

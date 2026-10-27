INSERT INTO work_task_labels (
    task_id, label_id, active, removed_at, created_at, updated_at
)
VALUES (
    $1, $2, $3,
    CASE WHEN $3 THEN NULL ELSE transaction_timestamp() END,
    transaction_timestamp(), transaction_timestamp()
)
ON CONFLICT (task_id, label_id) DO UPDATE
SET active = EXCLUDED.active,
    removed_at = CASE WHEN EXCLUDED.active THEN NULL ELSE transaction_timestamp() END,
    updated_at = transaction_timestamp()

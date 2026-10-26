UPDATE work_task_assignees
SET active = false, removed_at = transaction_timestamp(), updated_at = transaction_timestamp()
WHERE task_id = $1 AND active

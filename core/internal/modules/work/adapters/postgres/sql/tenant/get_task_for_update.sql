SELECT id, board_id, column_id, creator_user_id, task_kind, status_category, priority, title, description,
       position, row_version, starts_at, due_at, deleted_at
FROM work_tasks
WHERE id = $1
FOR UPDATE

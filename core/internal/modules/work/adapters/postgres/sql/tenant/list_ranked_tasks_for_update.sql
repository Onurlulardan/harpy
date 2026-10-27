SELECT id, position
FROM work_tasks
WHERE column_id = $1 AND deleted_at IS NULL
ORDER BY position, id
FOR UPDATE

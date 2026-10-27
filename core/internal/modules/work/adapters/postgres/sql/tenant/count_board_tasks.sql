SELECT count(*)
FROM work_tasks
WHERE board_id = $1 AND deleted_at IS NULL

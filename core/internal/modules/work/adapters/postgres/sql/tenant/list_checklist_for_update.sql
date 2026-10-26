SELECT id, task_id, item_text, completed, position, row_version
FROM work_checklist_items
WHERE task_id = $1 AND deleted_at IS NULL
ORDER BY position, id
FOR UPDATE

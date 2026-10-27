SELECT COALESCE(metadata.is_recurring, false),
       COALESCE(metadata.recurrence_rule, ''),
       metadata.recurrence_end,
       metadata.form_filter
FROM work_tasks AS task
LEFT JOIN work_task_compatibility_metadata AS metadata
  ON metadata.task_id = task.id
WHERE task.id = $1
FOR UPDATE OF task

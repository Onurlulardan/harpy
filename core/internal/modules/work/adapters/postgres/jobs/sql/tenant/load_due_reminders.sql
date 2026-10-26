SELECT task.id, task.row_version, task.due_at, task.title
FROM work_tasks task
WHERE task.due_at > $1 - interval '24 hours' AND task.due_at <= $1
  AND task.deleted_at IS NULL
  AND task.status_category NOT IN ('DONE','CANCELLED')
  AND NOT EXISTS (
      SELECT 1 FROM work_reminder_receipts receipt
      WHERE receipt.task_id = task.id
        AND receipt.task_version = task.row_version
        AND receipt.due_at = task.due_at
  )
ORDER BY task.due_at, task.id
FOR UPDATE OF task SKIP LOCKED
LIMIT $2

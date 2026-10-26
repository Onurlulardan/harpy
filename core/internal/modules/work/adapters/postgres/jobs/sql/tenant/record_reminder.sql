INSERT INTO work_reminder_receipts (
    task_id, task_version, due_at, event_id, dispatched_at
) VALUES ($1,$2,$3,$4,$5)
ON CONFLICT (task_id, task_version, due_at) DO NOTHING

INSERT INTO work_checklist_items (
    id, task_id, item_text, completed, position, row_version, deleted_at,
    created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,1,NULL,transaction_timestamp(),transaction_timestamp())
ON CONFLICT (id) DO UPDATE SET
    task_id=EXCLUDED.task_id, item_text=EXCLUDED.item_text,
    completed=EXCLUDED.completed, position=EXCLUDED.position,
    deleted_at=NULL, updated_at=EXCLUDED.updated_at

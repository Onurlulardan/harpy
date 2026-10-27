INSERT INTO work_task_labels (
    task_id, label_id, active, removed_at, created_at, updated_at
) VALUES ($1,$2,true,NULL,transaction_timestamp(),transaction_timestamp())
ON CONFLICT (task_id, label_id) DO UPDATE SET
    active=true, removed_at=NULL, updated_at=EXCLUDED.updated_at

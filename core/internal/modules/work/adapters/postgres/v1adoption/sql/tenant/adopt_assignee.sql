INSERT INTO work_task_assignees (
    task_id, principal_kind, principal_id, active, removed_at, created_at, updated_at
) VALUES ($1,$2,$3,true,NULL,transaction_timestamp(),transaction_timestamp())
ON CONFLICT (task_id, principal_kind, principal_id) DO UPDATE SET
    active=true, removed_at=NULL, updated_at=EXCLUDED.updated_at

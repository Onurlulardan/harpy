INSERT INTO work_tasks (
    id, board_id, column_id, creator_user_id, task_kind, status_category, priority,
    title, description, position, row_version, starts_at, due_at, completed_at, cancelled_at, deleted_at,
    created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,1,$11,$12,$13,$14,$15,$16,$17)
ON CONFLICT (id) DO UPDATE SET
    board_id=EXCLUDED.board_id, column_id=EXCLUDED.column_id,
    creator_user_id=EXCLUDED.creator_user_id,
    task_kind=EXCLUDED.task_kind, status_category=EXCLUDED.status_category,
    priority=EXCLUDED.priority, title=EXCLUDED.title, description=EXCLUDED.description,
    position=EXCLUDED.position, starts_at=EXCLUDED.starts_at, due_at=EXCLUDED.due_at,
    completed_at=EXCLUDED.completed_at, cancelled_at=EXCLUDED.cancelled_at,
    deleted_at=EXCLUDED.deleted_at, created_at=EXCLUDED.created_at,
    updated_at=EXCLUDED.updated_at

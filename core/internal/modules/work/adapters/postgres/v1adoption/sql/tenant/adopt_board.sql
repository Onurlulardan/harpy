INSERT INTO work_boards (
    id, creator_user_id, name, description, lifecycle_state, reference_epoch,
    row_version, created_at, updated_at, archived_at
) VALUES ($1,$2,$3,$4,$5,1,1,$6,$7,$8)
ON CONFLICT (id) DO UPDATE SET
    creator_user_id=EXCLUDED.creator_user_id, name=EXCLUDED.name,
    description=EXCLUDED.description, lifecycle_state=EXCLUDED.lifecycle_state,
    created_at=EXCLUDED.created_at, updated_at=EXCLUDED.updated_at,
    archived_at=EXCLUDED.archived_at

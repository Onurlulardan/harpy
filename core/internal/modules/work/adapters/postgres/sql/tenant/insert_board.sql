INSERT INTO work_boards (
    id, creator_user_id, name, description, lifecycle_state, reference_epoch,
    row_version, created_at, updated_at, archived_at
) VALUES (
    $1, $2, $3, $4, $5, 1, $6, $7, $7,
    CASE WHEN $5 = 'ARCHIVED' THEN $7 ELSE NULL END
)

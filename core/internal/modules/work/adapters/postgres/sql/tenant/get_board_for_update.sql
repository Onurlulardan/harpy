SELECT id, creator_user_id, name, description, lifecycle_state, row_version, created_at
FROM work_boards
WHERE id = $1
FOR UPDATE

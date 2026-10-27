SELECT id, board_id, name, color, row_version, deleted_at
FROM work_labels
WHERE id = $1
FOR UPDATE

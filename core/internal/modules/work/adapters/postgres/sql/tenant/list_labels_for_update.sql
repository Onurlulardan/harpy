SELECT id, board_id, name, color, row_version
FROM work_labels
WHERE board_id = $1 AND deleted_at IS NULL
ORDER BY id
FOR UPDATE

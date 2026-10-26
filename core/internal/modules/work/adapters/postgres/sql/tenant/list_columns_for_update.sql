SELECT id, board_id, name, color, status_category, position, row_version
FROM work_board_columns
WHERE board_id = $1
ORDER BY position, id
FOR UPDATE

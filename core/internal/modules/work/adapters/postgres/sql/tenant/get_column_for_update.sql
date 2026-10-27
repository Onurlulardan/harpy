SELECT id, board_id, name, color, status_category, position, row_version
FROM work_board_columns
WHERE id = $1
FOR UPDATE

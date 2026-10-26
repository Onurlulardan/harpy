UPDATE work_board_columns
SET name = $2, color = $3, status_category = $4, position = $5,
    row_version = $6, updated_at = transaction_timestamp()
WHERE id = $1 AND row_version = $7

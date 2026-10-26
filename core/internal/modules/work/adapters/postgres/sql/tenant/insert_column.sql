INSERT INTO work_board_columns (
    id, board_id, name, color, status_category, position, row_version, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, transaction_timestamp(), transaction_timestamp())

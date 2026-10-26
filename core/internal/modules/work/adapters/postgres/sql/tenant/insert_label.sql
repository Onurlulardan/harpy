INSERT INTO work_labels (id, board_id, name, color, row_version, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, transaction_timestamp(), transaction_timestamp())

UPDATE work_board_memberships
SET active = false, row_version = row_version + 1,
    updated_at = transaction_timestamp(), deactivated_at = transaction_timestamp()
WHERE id = $1 AND row_version = $2 AND active

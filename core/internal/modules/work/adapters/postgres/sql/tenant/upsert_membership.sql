INSERT INTO work_board_memberships (
    id, board_id, principal_kind, principal_id, board_role, active,
    row_version, created_at, updated_at, deactivated_at
) VALUES ($1, $2, $3, $4, $5, true, $6, transaction_timestamp(), transaction_timestamp(), NULL)
ON CONFLICT (id) DO UPDATE SET
    principal_kind = EXCLUDED.principal_kind,
    principal_id = EXCLUDED.principal_id,
    board_role = EXCLUDED.board_role,
    active = true,
    row_version = EXCLUDED.row_version,
    updated_at = transaction_timestamp(),
    deactivated_at = NULL
WHERE work_board_memberships.board_id = EXCLUDED.board_id
  AND work_board_memberships.row_version = EXCLUDED.row_version - 1

SELECT id, board_id, principal_kind, principal_id, board_role, active, row_version
FROM work_board_memberships
WHERE board_id = $1
ORDER BY id
FOR UPDATE

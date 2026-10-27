SELECT id, board_id, user_id, role,
       created_at AT TIME ZONE 'UTC', updated_at AT TIME ZONE 'UTC'
FROM board_members
WHERE tenant_id = $1
ORDER BY id

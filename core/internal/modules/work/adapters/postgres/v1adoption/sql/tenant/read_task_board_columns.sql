SELECT id, board_id, name, COALESCE(color, ''), COALESCE(linked_status, ''), "order",
       created_at AT TIME ZONE 'UTC', updated_at AT TIME ZONE 'UTC'
FROM task_board_columns
WHERE tenant_id = $1
ORDER BY id

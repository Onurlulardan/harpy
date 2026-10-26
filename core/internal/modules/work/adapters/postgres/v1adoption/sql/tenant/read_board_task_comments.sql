SELECT id, task_id, user_id, content,
       created_at AT TIME ZONE 'UTC', updated_at AT TIME ZONE 'UTC',
       deleted_at AT TIME ZONE 'UTC'
FROM board_task_comments
WHERE tenant_id = $1
ORDER BY id

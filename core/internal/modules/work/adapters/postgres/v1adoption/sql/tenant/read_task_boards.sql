SELECT id, organization_id, created_by, name, COALESCE(description, ''),
       created_at AT TIME ZONE 'UTC', updated_at AT TIME ZONE 'UTC',
       deleted_at AT TIME ZONE 'UTC'
FROM task_boards
WHERE tenant_id = $1
ORDER BY id

SELECT id, board_id, column_id, created_by, title, COALESCE(description, ''),
       task_type, priority, status, COALESCE(form_slug, ''), COALESCE(assignee_type, ''),
       column_order, form_data_id, assignee_id, workflow_task_id,
       due_date AT TIME ZONE 'UTC', start_date AT TIME ZONE 'UTC',
       completed_at AT TIME ZONE 'UTC', recurrence_end AT TIME ZONE 'UTC',
       deleted_at AT TIME ZONE 'UTC', COALESCE(is_recurring, false),
       COALESCE(recurrence_rule, ''), form_filter,
       COALESCE(labels, '[]'::jsonb), COALESCE(checklist, '[]'::jsonb),
       created_at AT TIME ZONE 'UTC', updated_at AT TIME ZONE 'UTC'
FROM board_tasks
WHERE tenant_id = $1
ORDER BY id

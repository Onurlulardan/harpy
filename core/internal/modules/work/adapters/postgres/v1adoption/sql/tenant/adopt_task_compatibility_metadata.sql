INSERT INTO work_task_compatibility_metadata (
    task_id, is_recurring, recurrence_rule, recurrence_end, form_filter,
    created_at, updated_at
) VALUES ($1, $2, $3, $4, $5::jsonb, $6, $7)
ON CONFLICT (task_id) DO UPDATE SET
    is_recurring = EXCLUDED.is_recurring,
    recurrence_rule = EXCLUDED.recurrence_rule,
    recurrence_end = EXCLUDED.recurrence_end,
    form_filter = EXCLUDED.form_filter,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at

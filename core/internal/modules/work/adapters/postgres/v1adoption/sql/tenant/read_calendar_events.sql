SELECT id, organization_id, created_by, title, COALESCE(description, ''), color,
       start_date AT TIME ZONE 'UTC', end_date AT TIME ZONE 'UTC',
       COALESCE(all_day, true), COALESCE(form_slug, ''), COALESCE(form_data_id, ''),
       created_at AT TIME ZONE 'UTC', updated_at AT TIME ZONE 'UTC',
       deleted_at AT TIME ZONE 'UTC'
FROM calendar_events
WHERE tenant_id = $1
ORDER BY id

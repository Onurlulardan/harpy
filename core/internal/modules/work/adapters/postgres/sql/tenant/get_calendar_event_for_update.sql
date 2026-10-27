SELECT id, owner_id, audience, board_id, organization_id, title, description, color,
       starts_at, ends_at, start_date, end_date, timezone_name, rule_id,
       rule_revision, generated, row_version, deleted_at
FROM work_calendar_events
WHERE id = $1
FOR UPDATE

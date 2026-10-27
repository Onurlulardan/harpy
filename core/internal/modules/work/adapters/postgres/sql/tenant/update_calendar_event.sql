UPDATE work_calendar_events
SET audience = $2, board_id = $3, organization_id = $4, title = $5, description = $6,
    color = $7, starts_at = $8, ends_at = $9, start_date = $10::date, end_date = $11::date,
    timezone_name = $12, row_version = $13, deleted_at = $14,
    updated_at = transaction_timestamp(), edited = true
WHERE id = $1 AND row_version = $15 AND NOT generated

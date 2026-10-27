INSERT INTO work_calendar_events (
    id, owner_id, audience, board_id, organization_id, title, description,
    starts_at, ends_at, start_date, end_date, timezone_name, rule_id, rule_revision,
    occurrence_start, tzdb_version, generated, generated_state, edited,
    live_reference_count, legal_hold, row_version, deleted_at, created_at, updated_at
)
SELECT $2, event.owner_id, event.audience, event.board_id, event.organization_id,
       event.title, event.description,
       CASE WHEN event.starts_at IS NULL THEN NULL ELSE $3::timestamptz END,
       CASE WHEN event.ends_at IS NULL THEN NULL ELSE $3::timestamptz + (event.ends_at - event.starts_at) END,
       CASE WHEN event.start_date IS NULL THEN NULL ELSE $5::date END,
       CASE WHEN event.end_date IS NULL THEN NULL ELSE $5::date + (event.end_date - event.start_date) END,
       event.timezone_name, rule.id, rule.current_revision, $3, $4, true, 'ACTIVE',
       false, 0, false, 1, NULL, transaction_timestamp(), transaction_timestamp()
FROM work_recurrence_rules rule
JOIN work_calendar_events event ON event.id = rule.owner_event_id
WHERE rule.id = $1 AND rule.lifecycle_state = 'ACTIVE' AND event.deleted_at IS NULL
ON CONFLICT (rule_id, rule_revision, occurrence_start) WHERE generated DO NOTHING

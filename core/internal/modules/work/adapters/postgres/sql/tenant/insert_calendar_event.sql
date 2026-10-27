INSERT INTO work_calendar_events (
    id, owner_id, audience, board_id, organization_id, title, description, color,
    starts_at, ends_at, start_date, end_date, timezone_name, rule_id, rule_revision,
    occurrence_start, tzdb_version, generated, generated_state, edited,
    live_reference_count, legal_hold, row_version, deleted_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::date, $12::date, $13,
    NULL, NULL, NULL, NULL, false, NULL, false, 0, false, $14, NULL,
    transaction_timestamp(), transaction_timestamp()
)

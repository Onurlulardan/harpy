INSERT INTO work_calendar_items (
    id, source_owner, source_kind, source_id, source_version_id, source_revision,
    permission_revision, event_sequence, summary, starts_at, ends_at, start_date,
    end_date, state, visible_to_principals, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::date,$13::date,$14,$15,transaction_timestamp())
ON CONFLICT (source_owner, source_kind, source_id, source_version_id) DO UPDATE SET
    source_revision = EXCLUDED.source_revision,
    permission_revision = EXCLUDED.permission_revision,
    event_sequence = EXCLUDED.event_sequence,
    summary = EXCLUDED.summary,
    starts_at = EXCLUDED.starts_at,
    ends_at = EXCLUDED.ends_at,
    start_date = EXCLUDED.start_date,
    end_date = EXCLUDED.end_date,
    state = EXCLUDED.state,
    visible_to_principals = EXCLUDED.visible_to_principals,
    updated_at = transaction_timestamp()
WHERE work_calendar_items.event_sequence <= EXCLUDED.event_sequence

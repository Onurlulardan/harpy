INSERT INTO work_inbox_items (
    id, source_owner, source_kind, source_id, source_version_id, source_revision,
    permission_revision, event_sequence, summary, due_at, state,
    visible_to_principals, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,transaction_timestamp())
ON CONFLICT (source_owner, source_kind, source_id, source_version_id) DO UPDATE SET
    source_revision = EXCLUDED.source_revision,
    permission_revision = EXCLUDED.permission_revision,
    event_sequence = EXCLUDED.event_sequence,
    summary = EXCLUDED.summary,
    due_at = EXCLUDED.due_at,
    state = EXCLUDED.state,
    visible_to_principals = EXCLUDED.visible_to_principals,
    updated_at = transaction_timestamp()
WHERE work_inbox_items.event_sequence <= EXCLUDED.event_sequence

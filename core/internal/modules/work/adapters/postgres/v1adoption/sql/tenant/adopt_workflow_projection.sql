INSERT INTO work_inbox_items (
    id, source_owner, source_kind, source_id, source_version_id,
    source_revision, permission_revision, event_sequence, summary, due_at,
    state, visible_to_principals, updated_at
) VALUES ($1,'WORKFLOW','HUMAN_TASK',$2,$3,$4,$5,$4,$6,$7,'VISIBLE',$8,
    transaction_timestamp())
ON CONFLICT (id) DO UPDATE SET
    source_version_id=EXCLUDED.source_version_id,
    source_revision=EXCLUDED.source_revision,
    permission_revision=EXCLUDED.permission_revision,
    event_sequence=EXCLUDED.event_sequence, summary=EXCLUDED.summary,
    due_at=EXCLUDED.due_at, state='VISIBLE',
    visible_to_principals=EXCLUDED.visible_to_principals,
    updated_at=EXCLUDED.updated_at

INSERT INTO work_recurrence_rules (
    id, owner_event_id, current_revision, lifecycle_state, row_version,
    created_at, updated_at
) VALUES ($1,$2,1,'ACTIVE',1,transaction_timestamp(),transaction_timestamp())
ON CONFLICT (id) DO UPDATE SET
    owner_event_id=EXCLUDED.owner_event_id, current_revision=1,
    lifecycle_state='ACTIVE', updated_at=EXCLUDED.updated_at

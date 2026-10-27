INSERT INTO notifications_inbox_counters (recipient_id, current_revision)
VALUES ($1::uuid, $2)
ON CONFLICT (recipient_id) DO UPDATE
SET current_revision = GREATEST(
    notifications_inbox_counters.current_revision, EXCLUDED.current_revision
);

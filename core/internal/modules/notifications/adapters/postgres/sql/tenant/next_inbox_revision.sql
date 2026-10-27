INSERT INTO notifications_inbox_counters (recipient_id, current_revision)
VALUES ($1::uuid, 1)
ON CONFLICT (recipient_id) DO UPDATE
SET current_revision = notifications_inbox_counters.current_revision + 1
RETURNING current_revision;

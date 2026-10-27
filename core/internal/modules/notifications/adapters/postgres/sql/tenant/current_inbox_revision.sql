SELECT current_revision
FROM notifications_inbox_counters
WHERE recipient_id = $1::uuid;

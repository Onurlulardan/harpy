SELECT pg_catalog.count(*)
FROM notifications_messages
WHERE recipient_id = $1::uuid AND read_at IS NULL AND dismissed_at IS NULL;

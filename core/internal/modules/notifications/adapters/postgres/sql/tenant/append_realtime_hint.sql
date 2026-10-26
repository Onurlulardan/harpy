INSERT INTO notifications_realtime_hints (
    id, message_id, recipient_id, inbox_revision, created_at, published_at
) VALUES (
    pg_catalog.uuidv7(), $1::uuid, $2::uuid, $3,
    pg_catalog.transaction_timestamp(), NULL
)
ON CONFLICT (message_id) DO NOTHING;

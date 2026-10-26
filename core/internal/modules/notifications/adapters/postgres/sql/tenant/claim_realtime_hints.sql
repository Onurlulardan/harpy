WITH candidates AS (
    SELECT id
    FROM notifications_realtime_hints
    WHERE published_at IS NULL
    ORDER BY created_at, id
    LIMIT $1
    FOR UPDATE SKIP LOCKED
), claimed AS (
    UPDATE notifications_realtime_hints hint
    SET published_at = pg_catalog.transaction_timestamp()
    FROM candidates
    WHERE hint.id = candidates.id
    RETURNING hint.message_id, hint.recipient_id, hint.inbox_revision,
              hint.created_at, hint.id
)
SELECT message_id::text, recipient_id::text, inbox_revision
FROM claimed
ORDER BY created_at, id;

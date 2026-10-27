WITH updated AS (
    UPDATE notifications_messages
    SET read_at = pg_catalog.transaction_timestamp(), row_version = row_version + 1
    WHERE recipient_id = $1::uuid AND inbox_revision <= $2
      AND read_at IS NULL AND dismissed_at IS NULL
    RETURNING id
)
SELECT pg_catalog.count(*) FROM updated;

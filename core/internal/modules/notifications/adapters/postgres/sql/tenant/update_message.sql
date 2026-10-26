UPDATE notifications_messages
SET subject = $4, body = $5, row_version = $6, read_at = $7,
    dismissed_at = $8, purged_at = $9
WHERE id = $1::uuid AND recipient_id = $2::uuid AND row_version = $3;

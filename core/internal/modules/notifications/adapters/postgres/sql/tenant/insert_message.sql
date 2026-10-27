INSERT INTO notifications_messages (
    id, recipient_id, template_version_id, producer, notification_type,
    category, template_kind, locale, subject, body, data_digest, subject_digest,
    body_digest, row_version, inbox_revision, created_at, read_at, dismissed_at, purged_at
) VALUES (
    $1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19
);

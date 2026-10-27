INSERT INTO notifications_messages (
    id, recipient_id, template_version_id, producer, notification_type,
    category, template_kind, locale, subject, body, data_digest,
    subject_digest, body_digest, row_version, inbox_revision, created_at,
    read_at, dismissed_at, purged_at
) VALUES (
    $1::uuid, $2::uuid, '12000000-0000-4000-8000-000000000002',
    'legacy_v1', $3, $4, 'plain_text', 'tr', $5, $6,
    $7::bytea, $8::bytea, $9::bytea, 1, $10, $11, $12, NULL, NULL
);

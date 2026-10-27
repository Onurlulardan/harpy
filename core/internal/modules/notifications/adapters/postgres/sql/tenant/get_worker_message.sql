SELECT id::text, recipient_id::text, template_version_id::text,
       producer, notification_type, category, template_kind, locale, subject, body,
       data_digest, subject_digest, body_digest, row_version, inbox_revision,
       created_at, read_at, dismissed_at, purged_at
FROM notifications_messages
WHERE id = $1::uuid;

SELECT id::text, recipient_id::text, template_version_id::text,
       producer, notification_type, category, template_kind, locale, subject, body,
       data_digest, subject_digest, body_digest, row_version, inbox_revision,
       created_at, read_at, dismissed_at, purged_at
FROM notifications_messages
WHERE recipient_id = $1::uuid AND dismissed_at IS NULL
  AND (NOT $2::boolean OR read_at IS NULL)
  AND ($3::text = '' OR category = $3)
  AND ($4::bigint = 0 OR inbox_revision < $4)
ORDER BY inbox_revision DESC, id DESC
LIMIT $5;

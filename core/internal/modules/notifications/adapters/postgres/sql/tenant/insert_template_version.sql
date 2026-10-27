INSERT INTO notifications_template_versions (
    id, template_id, version_number, locale, template_kind, subject_pattern,
    body_pattern, variables, mandatory_email, content_digest, published_at
) VALUES (
    $1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8::jsonb, $9, $10,
    pg_catalog.transaction_timestamp()
);

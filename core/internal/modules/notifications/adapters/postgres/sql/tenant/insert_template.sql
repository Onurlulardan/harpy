INSERT INTO notifications_templates (
    id, notification_type, published_version_id, row_version, created_at, updated_at
) VALUES (
    $1::uuid, $2, NULL, $3,
    pg_catalog.transaction_timestamp(), pg_catalog.transaction_timestamp()
);

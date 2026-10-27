INSERT INTO notifications_smtp_settings_receipts (
    idempotency_key, request_digest, expected_settings_revision,
    result_settings_revision, provider_config_id, provider_revision, created_at
) VALUES (
    $1, $2, $3, $4, $5::uuid, $6,
    pg_catalog.transaction_timestamp()
);

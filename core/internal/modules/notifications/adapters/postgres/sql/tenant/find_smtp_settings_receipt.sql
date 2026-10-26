SELECT request_digest, expected_settings_revision, result_settings_revision,
       provider_config_id::text, provider_revision
FROM notifications_smtp_settings_receipts
WHERE idempotency_key = $1;

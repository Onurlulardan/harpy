SELECT operation_id::text, provider_config_version_id::text, actor_id::text,
       request_digest, expected_settings_revision, requested_at
FROM notifications_smtp_test_intents
WHERE idempotency_key = $1;

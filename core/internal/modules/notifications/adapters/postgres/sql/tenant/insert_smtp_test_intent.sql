INSERT INTO notifications_smtp_test_intents (
    operation_id, idempotency_key, request_digest, expected_settings_revision,
    provider_config_version_id, actor_id, target_ciphertext, target_nonce,
    target_wrapped_dek, target_kek_revision, target_digest, requested_at, created_at
) VALUES (
    $1::uuid, $2, $3, $4, $5::uuid, $6::uuid, $7, $8, $9, $10, $11, $12,
    pg_catalog.transaction_timestamp()
);

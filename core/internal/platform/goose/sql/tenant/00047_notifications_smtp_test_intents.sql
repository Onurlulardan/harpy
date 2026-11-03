-- +goose Up
CREATE TABLE notifications_smtp_test_intents (
    operation_id uuid NOT NULL PRIMARY KEY,
    idempotency_key text NOT NULL,
    request_digest text NOT NULL,
    expected_settings_revision bigint NOT NULL,
    provider_config_version_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    target_ciphertext text NOT NULL,
    target_nonce text NOT NULL,
    target_wrapped_dek text NOT NULL,
    target_kek_revision bigint NOT NULL,
    target_digest text NOT NULL,
    requested_at timestamptz,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE notifications_smtp_test_intents;

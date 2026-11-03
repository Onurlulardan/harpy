-- +goose Up
CREATE TABLE notifications_smtp_settings_receipts (
    idempotency_key text NOT NULL PRIMARY KEY,
    request_digest text NOT NULL,
    expected_settings_revision bigint NOT NULL,
    result_settings_revision bigint NOT NULL,
    provider_config_id uuid NOT NULL,
    provider_revision bigint NOT NULL,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE notifications_smtp_settings_receipts;

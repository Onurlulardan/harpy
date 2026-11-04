-- +goose Up
CREATE TABLE platform.notification_provider_locators (
    provider_kind text NOT NULL PRIMARY KEY,
    provider_account_digest text NOT NULL,
    external_message_digest text NOT NULL,
    tenant_id uuid NOT NULL,
    delivery_id uuid NOT NULL,
    accepted_at timestamptz,
    expires_at timestamptz
);

-- +goose Down
DROP TABLE platform.notification_provider_locators;

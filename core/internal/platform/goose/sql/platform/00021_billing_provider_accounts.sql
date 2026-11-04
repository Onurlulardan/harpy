-- +goose Up
CREATE TABLE platform.billing_provider_accounts (
    id uuid NOT NULL PRIMARY KEY,
    provider_key text NOT NULL,
    provider_account_key text NOT NULL,
    lifecycle text NOT NULL,
    supports_webhooks text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_provider_accounts;

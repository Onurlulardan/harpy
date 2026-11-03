-- +goose Up
CREATE TABLE platform.billing_subscriptions (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    state text NOT NULL,
    offering_version_id uuid NOT NULL,
    trial_ends_at timestamptz,
    period_end text NOT NULL,
    grace_ends_at timestamptz,
    price_version_id uuid NOT NULL,
    revision bigint NOT NULL,
    provider_account_id uuid NOT NULL,
    period_start text NOT NULL,
    cancel_at_period_end text NOT NULL,
    cancel_at timestamptz,
    updated_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_subscriptions;

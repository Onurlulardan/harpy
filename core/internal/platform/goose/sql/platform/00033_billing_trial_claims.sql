-- +goose Up
CREATE TABLE platform.billing_trial_claims (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    offering_version_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    claimed_at timestamptz
);

-- +goose Down
DROP TABLE platform.billing_trial_claims;

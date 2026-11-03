-- +goose Up
CREATE TABLE platform.billing_subscription_operations (
    id uuid NOT NULL PRIMARY KEY,
    subscription_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    source_kind text NOT NULL,
    operation_kind text NOT NULL,
    request_sha256 text NOT NULL,
    result_sha256 text NOT NULL,
    result_provider_account_id uuid NOT NULL,
    result_offering_version_id uuid NOT NULL,
    result_price_version_id uuid NOT NULL,
    result_currency text NOT NULL,
    result_state text NOT NULL,
    result_period_start text NOT NULL,
    result_period_end text NOT NULL,
    result_trial_ends_at timestamptz,
    result_grace_ends_at timestamptz,
    result_cancel_at_period_end text NOT NULL,
    result_cancel_at timestamptz,
    result_revision bigint NOT NULL,
    result_updated_at timestamptz,
    completed_at timestamptz,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    expected_revision bigint NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_subscription_operations;

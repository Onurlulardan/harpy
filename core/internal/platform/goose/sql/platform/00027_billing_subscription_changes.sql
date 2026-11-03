-- +goose Up
CREATE TABLE platform.billing_subscription_changes (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    before_offering_version_id uuid NOT NULL,
    before_price_version_id uuid NOT NULL,
    after_offering_version_id uuid NOT NULL,
    after_price_version_id uuid NOT NULL,
    source text NOT NULL,
    state text NOT NULL,
    effective_at timestamptz,
    source_operation_id uuid NOT NULL,
    decision_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    resolved_at timestamptz,
    revision bigint NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_subscription_changes;

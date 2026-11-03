-- +goose Up
CREATE TABLE platform.billing_entitlement_decisions (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    entitlement_definition_id uuid NOT NULL,
    entitlement_key text NOT NULL,
    meter_period_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    requested_amount text NOT NULL,
    current_amount text NOT NULL,
    limit_amount text NOT NULL,
    unlimited text NOT NULL,
    allowed text NOT NULL,
    soft_exceeded text NOT NULL,
    reason text NOT NULL,
    correlation_id uuid NOT NULL,
    decided_at timestamptz,
    retain_until text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_entitlement_decisions;

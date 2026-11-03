-- +goose Up
CREATE TABLE platform.billing_usage_reservations (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    decision_id uuid NOT NULL,
    entitlement_key text NOT NULL,
    meter_period_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    amount text NOT NULL,
    state text NOT NULL,
    expires_at timestamptz,
    effect_started_at timestamptz,
    effect_receipt_id uuid NOT NULL,
    resolution_receipt_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    resolved_at timestamptz,
    revision bigint NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_usage_reservations;

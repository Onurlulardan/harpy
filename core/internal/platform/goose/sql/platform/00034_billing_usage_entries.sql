-- +goose Up
CREATE TABLE platform.billing_usage_entries (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    decision_id uuid NOT NULL,
    entitlement_key text NOT NULL,
    meter_period_id uuid NOT NULL,
    reservation_id uuid NOT NULL,
    entry_kind text NOT NULL,
    amount_delta text NOT NULL,
    committed_after text NOT NULL,
    reserved_after text NOT NULL,
    gauge_after text NOT NULL,
    source_operation_id uuid NOT NULL,
    owner_fact_id uuid NOT NULL,
    evidence_sha256 text NOT NULL,
    occurred_at timestamptz,
    created_at timestamptz NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_usage_entries;

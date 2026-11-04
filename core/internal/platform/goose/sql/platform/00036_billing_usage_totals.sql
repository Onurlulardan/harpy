-- +goose Up
CREATE TABLE platform.billing_usage_totals (
    reserved_total text NOT NULL PRIMARY KEY,
    revision bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    meter_period_id uuid NOT NULL,
    entitlement_key text NOT NULL,
    meter_kind text NOT NULL,
    committed_total text NOT NULL,
    gauge_total text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_usage_totals;

-- +goose Up
CREATE TABLE platform.billing_meter_periods (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    entitlement_definition_id uuid NOT NULL,
    entitlement_key text NOT NULL,
    period_start text NOT NULL,
    period_end text NOT NULL,
    state text NOT NULL,
    meter_kind text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_meter_periods;

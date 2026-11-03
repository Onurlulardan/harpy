-- +goose Up
CREATE TABLE platform.billing_entitlement_definitions (
    id uuid NOT NULL PRIMARY KEY,
    entitlement_key text NOT NULL,
    value_type text NOT NULL,
    meter_kind text NOT NULL,
    reset_policy text NOT NULL,
    enforcement text NOT NULL,
    unlimited_allowed text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_entitlement_definitions;

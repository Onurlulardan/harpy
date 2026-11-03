-- +goose Up
CREATE TABLE platform.billing_subscription_overrides (
    id uuid NOT NULL PRIMARY KEY,
    integer_value text NOT NULL,
    unlimited text NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    entitlement_definition_id uuid NOT NULL,
    effective_start text NOT NULL,
    effective_end text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_subscription_overrides;

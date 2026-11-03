-- +goose Up
CREATE TABLE platform.billing_offering_entitlements (
    entitlement_definition_id uuid NOT NULL PRIMARY KEY,
    integer_value text NOT NULL,
    unlimited text NOT NULL,
    offering_version_id uuid NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_offering_entitlements;

-- +goose Up
CREATE TABLE platform.billing_discount_redemptions (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    discount_version_id uuid NOT NULL,
    offering_version_id uuid NOT NULL,
    price_version_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    applied_kind text NOT NULL,
    applied_fixed_minor_units text NOT NULL,
    applied_currency text NOT NULL,
    applied_currency_registry_version bigint NOT NULL,
    applied_basis_points text NOT NULL,
    applied_priority text NOT NULL,
    applied_duration text NOT NULL,
    applied_repeating_cycles text NOT NULL,
    policy_sha256 text NOT NULL,
    redeemed_at timestamptz,
    relationship_ended_at timestamptz
);

-- +goose Down
DROP TABLE platform.billing_discount_redemptions;

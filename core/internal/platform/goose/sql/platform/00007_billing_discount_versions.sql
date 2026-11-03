-- +goose Up
CREATE TABLE platform.billing_discount_versions (
    id uuid NOT NULL PRIMARY KEY,
    discount_kind text NOT NULL,
    fixed_minor_units text NOT NULL,
    currency text NOT NULL,
    currency_registry_version bigint NOT NULL,
    basis_points text NOT NULL,
    priority bigint NOT NULL,
    duration text NOT NULL,
    repeating_cycles text NOT NULL,
    valid_start text NOT NULL,
    valid_end text NOT NULL,
    global_use_limit text NOT NULL,
    per_tenant_use_limit text NOT NULL,
    all_offerings text NOT NULL,
    version bigint NOT NULL,
    available boolean NOT NULL,
    discount_id uuid NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_discount_versions;

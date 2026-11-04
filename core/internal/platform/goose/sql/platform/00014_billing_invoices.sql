-- +goose Up
CREATE TABLE platform.billing_invoices (
    id uuid NOT NULL PRIMARY KEY,
    revision bigint NOT NULL,
    state text NOT NULL,
    currency text NOT NULL,
    currency_registry_version bigint NOT NULL,
    settled_minor_units text NOT NULL,
    policy_sha256 text NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_invoices;

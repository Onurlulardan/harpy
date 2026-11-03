-- +goose Up
CREATE TABLE platform.billing_tenant_authorities (
    authority_id uuid NOT NULL PRIMARY KEY,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_group_id uuid NOT NULL,
    billing_domain text NOT NULL,
    tenant_id uuid NOT NULL,
    authority_group_generation bigint NOT NULL,
    billing_generation bigint NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_tenant_authorities;

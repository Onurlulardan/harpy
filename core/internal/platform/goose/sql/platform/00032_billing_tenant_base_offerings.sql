-- +goose Up
CREATE TABLE platform.billing_tenant_base_offerings (
    offering_version_id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    effective_start text NOT NULL,
    effective_end text NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_tenant_base_offerings;

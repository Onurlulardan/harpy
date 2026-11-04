-- +goose Up
CREATE TABLE platform.billing_global_authorities (
    authority_id uuid NOT NULL PRIMARY KEY,
    scope_kind text NOT NULL,
    scope_id uuid NOT NULL,
    authority_group_id uuid NOT NULL,
    authority_domain text NOT NULL,
    authority_group_generation bigint NOT NULL,
    authority_domain_generation bigint NOT NULL
);

-- +goose Down
DROP TABLE platform.billing_global_authorities;

-- +goose Up
CREATE TABLE platform.billing_provider_objects (
    tenant_id uuid NOT NULL,
    id uuid NOT NULL PRIMARY KEY,
    subscription_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    revision bigint NOT NULL,
    provider_account_id uuid NOT NULL,
    authority_id uuid NOT NULL,
    object_kind text NOT NULL,
    provider_object_id uuid NOT NULL,
    retired_at timestamptz
);

-- +goose Down
DROP TABLE platform.billing_provider_objects;

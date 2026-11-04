-- +goose Up
CREATE TABLE platform.billing_invoice_finalizations (
    id uuid NOT NULL PRIMARY KEY,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    expected_revision bigint NOT NULL,
    result_revision bigint NOT NULL,
    invoice_total_minor_units text NOT NULL,
    settled_minor_units text NOT NULL,
    amount_due_minor_units text NOT NULL,
    credit_balance_minor_units text NOT NULL,
    currency text NOT NULL,
    currency_registry_version bigint NOT NULL,
    policy_sha256 text NOT NULL,
    finalized_at timestamptz
);

-- +goose Down
DROP TABLE platform.billing_invoice_finalizations;

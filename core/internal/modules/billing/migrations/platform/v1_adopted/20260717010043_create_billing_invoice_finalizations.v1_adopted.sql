-- +goose Up
CREATE TABLE platform.billing_invoice_finalizations (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    expected_revision bigint NOT NULL,
    result_revision bigint NOT NULL,
    invoice_total_minor_units bigint NOT NULL,
    settled_minor_units bigint NOT NULL,
    amount_due_minor_units bigint NOT NULL,
    credit_balance_minor_units bigint NOT NULL,
    currency character(3) COLLATE "C" NOT NULL,
    currency_registry_version bigint NOT NULL,
    policy_sha256 bytea NOT NULL,
    finalized_at timestamptz NOT NULL,
    CONSTRAINT billing_invoice_finalizations_pkey PRIMARY KEY (id),
    CONSTRAINT billing_invoice_finalizations_invoice_key UNIQUE (tenant_id, invoice_id),
    CONSTRAINT billing_invoice_finalizations_operation_key
        UNIQUE (tenant_id, source_operation_id),
    CONSTRAINT billing_invoice_finalizations_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_invoice_finalizations_invoice_fk FOREIGN KEY (
        invoice_id, tenant_id, currency, currency_registry_version
    ) REFERENCES platform.billing_invoices (
        id, tenant_id, currency, currency_registry_version
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_invoice_finalizations_shape_check CHECK (
        expected_revision > 0 AND result_revision = expected_revision + 1
        AND pg_catalog.abs(invoice_total_minor_units) <= 9000000000000000
        AND settled_minor_units BETWEEN 0 AND 9000000000000000
        AND amount_due_minor_units BETWEEN 0 AND 9000000000000000
        AND credit_balance_minor_units BETWEEN 0 AND 9000000000000000
        AND amount_due_minor_units::numeric = GREATEST(
            0, invoice_total_minor_units::numeric - settled_minor_units::numeric
        )
        AND credit_balance_minor_units::numeric = GREATEST(
            0, settled_minor_units::numeric - invoice_total_minor_units::numeric
        )
        AND pg_catalog.octet_length(policy_sha256) = 32
    )
);

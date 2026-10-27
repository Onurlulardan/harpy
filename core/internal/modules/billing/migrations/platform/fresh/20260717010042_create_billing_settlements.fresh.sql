-- +goose Up
CREATE TABLE platform.billing_settlements (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    provider_account_id uuid NOT NULL,
    settlement_kind text NOT NULL,
    minor_units bigint NOT NULL,
    currency character(3) COLLATE "C" NOT NULL,
    currency_registry_version bigint NOT NULL,
    provider_receipt_reference text COLLATE "C" NOT NULL,
    source_operation_id uuid NOT NULL,
    occurred_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_settlements_pkey PRIMARY KEY (id),
    CONSTRAINT billing_settlements_provider_receipt_key UNIQUE (
        provider_account_id, provider_receipt_reference
    ),
    CONSTRAINT billing_settlements_operation_key
        UNIQUE (tenant_id, source_operation_id),
    CONSTRAINT billing_settlements_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_settlements_invoice_fk FOREIGN KEY (
        invoice_id, tenant_id, currency, currency_registry_version
    ) REFERENCES platform.billing_invoices (
        id, tenant_id, currency, currency_registry_version
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_settlements_provider_fk FOREIGN KEY (provider_account_id)
        REFERENCES platform.billing_provider_accounts (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_settlements_shape_check CHECK (
        settlement_kind IN ('paid', 'refunded')
        AND minor_units BETWEEN 1 AND 9000000000000000
        AND pg_catalog.octet_length(provider_receipt_reference) BETWEEN 1 AND 512
        AND created_at >= occurred_at
    )
);

CREATE INDEX billing_settlements_invoice_time_idx
    ON platform.billing_settlements (tenant_id, invoice_id, occurred_at, id);


-- +goose Up
CREATE TABLE platform.billing_invoice_lines (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    line_sequence integer NOT NULL,
    line_kind text NOT NULL,
    description text NOT NULL,
    quantity bigint NOT NULL,
    unit_minor_units bigint NOT NULL,
    gross_minor_units bigint NOT NULL,
    discount_minor_units bigint NOT NULL,
    tax_minor_units bigint NOT NULL,
    debit_adjustment_minor_units bigint NOT NULL,
    credit_adjustment_minor_units bigint NOT NULL,
    line_total_minor_units bigint NOT NULL,
    currency character(3) COLLATE "C" NOT NULL,
    currency_registry_version bigint NOT NULL,
    allocation_rank integer NOT NULL,
    money_policy_version text NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_invoice_lines_pkey PRIMARY KEY (id),
    CONSTRAINT billing_invoice_lines_sequence_key UNIQUE (invoice_id, line_sequence),
    CONSTRAINT billing_invoice_lines_parent_key UNIQUE (tenant_id, invoice_id, id),
    CONSTRAINT billing_invoice_lines_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_invoice_lines_invoice_fk FOREIGN KEY (
        invoice_id, tenant_id, currency, currency_registry_version
    ) REFERENCES platform.billing_invoices (
        id, tenant_id, currency, currency_registry_version
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_invoice_lines_shape_check CHECK (
        line_sequence > 0 AND allocation_rank > 0
        AND line_kind IN ('item', 'tax', 'discount', 'adjustment', 'credit')
        AND pg_catalog.octet_length(description) BETWEEN 1 AND 2048
        AND quantity > 0 AND unit_minor_units >= 0
        AND money_policy_version = 'fc_money_v1'
    ),
    CONSTRAINT billing_invoice_lines_amount_check CHECK (
        gross_minor_units BETWEEN 0 AND 9000000000000000
        AND discount_minor_units BETWEEN 0 AND gross_minor_units
        AND tax_minor_units BETWEEN 0 AND 9000000000000000
        AND debit_adjustment_minor_units BETWEEN 0 AND 9000000000000000
        AND credit_adjustment_minor_units BETWEEN 0 AND 9000000000000000
        AND pg_catalog.abs(line_total_minor_units) <= 9000000000000000
    ),
    CONSTRAINT billing_invoice_lines_total_check CHECK (
        gross_minor_units::numeric = unit_minor_units::numeric * quantity::numeric
        AND line_total_minor_units::numeric = gross_minor_units::numeric
            - discount_minor_units::numeric + tax_minor_units::numeric
            + debit_adjustment_minor_units::numeric
            - credit_adjustment_minor_units::numeric
    )
);

CREATE TABLE platform.billing_adjustments (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    invoice_id uuid NOT NULL,
    invoice_line_id uuid NOT NULL,
    adjustment_kind text NOT NULL,
    minor_units bigint NOT NULL,
    reason text NOT NULL,
    source_operation_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_adjustments_pkey PRIMARY KEY (id),
    CONSTRAINT billing_adjustments_operation_key
        UNIQUE (tenant_id, source_operation_id),
    CONSTRAINT billing_adjustments_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_adjustments_line_fk FOREIGN KEY (
        tenant_id, invoice_id, invoice_line_id
    ) REFERENCES platform.billing_invoice_lines (tenant_id, invoice_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_adjustments_shape_check CHECK (
        adjustment_kind IN ('credit', 'debit')
        AND minor_units BETWEEN 1 AND 9000000000000000
        AND pg_catalog.octet_length(reason) BETWEEN 1 AND 2048
    )
);


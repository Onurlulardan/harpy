-- +goose Up
CREATE TABLE platform.billing_invoices (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    invoice_number text COLLATE "C" NOT NULL,
    state text NOT NULL,
    currency character(3) COLLATE "C" NOT NULL,
    currency_registry_version bigint NOT NULL,
    currency_eligible boolean NOT NULL,
    gross_minor_units bigint NOT NULL,
    discount_minor_units bigint NOT NULL,
    tax_minor_units bigint NOT NULL,
    debit_adjustment_minor_units bigint NOT NULL,
    credit_adjustment_minor_units bigint NOT NULL,
    invoice_total_minor_units bigint NOT NULL,
    settled_minor_units bigint NOT NULL,
    amount_due_minor_units bigint NOT NULL,
    credit_balance_minor_units bigint NOT NULL,
    money_policy_version text NOT NULL,
    policy_sha256 bytea NOT NULL,
    period_start timestamptz NOT NULL,
    period_end timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    finalized_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_invoices_pkey PRIMARY KEY (id),
    CONSTRAINT billing_invoices_number_key UNIQUE (tenant_id, invoice_number),
    CONSTRAINT billing_invoices_tenant_id_key UNIQUE (tenant_id, id),
    CONSTRAINT billing_invoices_currency_key UNIQUE (
        id, tenant_id, currency, currency_registry_version
    ),
    CONSTRAINT billing_invoices_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_invoices_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_invoices_currency_fk FOREIGN KEY (
        currency_registry_version, currency, currency_eligible
    ) REFERENCES platform.billing_currency_exponents (
        registry_version, code, financial_money_eligible
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_invoices_state_check CHECK (
        (state = 'draft' AND finalized_at IS NULL)
        OR (state = 'finalized' AND finalized_at IS NOT NULL)
    ),
    CONSTRAINT billing_invoices_money_check CHECK (
        currency_eligible AND money_policy_version = 'fc_money_v1'
        AND pg_catalog.octet_length(policy_sha256) = 32
        AND gross_minor_units BETWEEN 0 AND 9000000000000000
        AND discount_minor_units BETWEEN 0 AND gross_minor_units
        AND tax_minor_units BETWEEN 0 AND 9000000000000000
        AND debit_adjustment_minor_units BETWEEN 0 AND 9000000000000000
        AND credit_adjustment_minor_units BETWEEN 0 AND 9000000000000000
        AND pg_catalog.abs(invoice_total_minor_units) <= 9000000000000000
        AND settled_minor_units BETWEEN 0 AND 9000000000000000
        AND amount_due_minor_units BETWEEN 0 AND 9000000000000000
        AND credit_balance_minor_units BETWEEN 0 AND 9000000000000000
    ),
    CONSTRAINT billing_invoices_total_check CHECK (
        invoice_total_minor_units::numeric =
            gross_minor_units::numeric - discount_minor_units::numeric
            + tax_minor_units::numeric + debit_adjustment_minor_units::numeric
            - credit_adjustment_minor_units::numeric
        AND amount_due_minor_units::numeric = GREATEST(
            0, invoice_total_minor_units::numeric - settled_minor_units::numeric
        )
        AND credit_balance_minor_units::numeric = GREATEST(
            0, settled_minor_units::numeric - invoice_total_minor_units::numeric
        )
    ),
    CONSTRAINT billing_invoices_time_check CHECK (
        period_start < period_end AND updated_at >= created_at
        AND (finalized_at IS NULL OR finalized_at >= created_at)
        AND revision > 0
    )
);

CREATE INDEX billing_invoices_tenant_state_idx
    ON platform.billing_invoices (tenant_id, state, created_at, id);

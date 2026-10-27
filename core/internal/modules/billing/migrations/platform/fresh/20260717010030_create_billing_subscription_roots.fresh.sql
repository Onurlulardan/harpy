-- +goose Up
CREATE TABLE platform.billing_tenant_base_offerings (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    offering_version_id uuid NOT NULL,
    effective_start timestamptz NOT NULL,
    effective_end timestamptz,
    source_operation_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_tenant_base_offerings_pkey PRIMARY KEY (id),
    CONSTRAINT billing_tenant_base_offerings_start_key
        UNIQUE (tenant_id, effective_start),
    CONSTRAINT billing_tenant_base_offerings_operation_key
        UNIQUE (tenant_id, source_operation_id),
    CONSTRAINT billing_tenant_base_offerings_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_tenant_base_offerings_offering_fk
        FOREIGN KEY (offering_version_id)
        REFERENCES platform.billing_offering_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_tenant_base_offerings_interval_check CHECK (
        effective_end IS NULL OR effective_start < effective_end
    )
);

CREATE TABLE platform.billing_subscriptions (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    provider_account_id uuid NOT NULL,
    offering_version_id uuid NOT NULL,
    price_version_id uuid NOT NULL,
    state text NOT NULL,
    period_start timestamptz,
    period_end timestamptz,
    trial_ends_at timestamptz,
    grace_ends_at timestamptz,
    cancel_at_period_end boolean NOT NULL,
    cancel_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    revision bigint NOT NULL,
    CONSTRAINT billing_subscriptions_pkey PRIMARY KEY (id),
    CONSTRAINT billing_subscriptions_tenant_id_key UNIQUE (tenant_id, id),
    CONSTRAINT billing_subscriptions_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscriptions_provider_fk FOREIGN KEY (provider_account_id)
        REFERENCES platform.billing_provider_accounts (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscriptions_offering_fk FOREIGN KEY (offering_version_id)
        REFERENCES platform.billing_offering_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscriptions_price_fk FOREIGN KEY (price_version_id)
        REFERENCES platform.billing_price_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscriptions_state_check CHECK (
        state IN (
            'pending', 'trialing', 'active', 'past_due',
            'paused', 'cancelled', 'expired'
        )
    ),
    CONSTRAINT billing_subscriptions_period_check CHECK (
        (period_start IS NULL AND period_end IS NULL)
        OR (period_start IS NOT NULL AND period_end IS NOT NULL
            AND period_start < period_end)
    ),
    CONSTRAINT billing_subscriptions_boundary_check CHECK (
        (state <> 'trialing' OR trial_ends_at IS NOT NULL)
        AND (state <> 'past_due' OR grace_ends_at IS NOT NULL)
        AND (cancel_at_period_end = (cancel_at IS NOT NULL))
        AND (cancel_at IS NULL OR period_end = cancel_at)
    ),
    CONSTRAINT billing_subscriptions_revision_check CHECK (
        revision > 0 AND updated_at >= created_at
    )
);

CREATE UNIQUE INDEX billing_subscriptions_one_grant_candidate_idx
    ON platform.billing_subscriptions (tenant_id)
    WHERE state IN ('trialing', 'active', 'past_due');
CREATE INDEX billing_subscriptions_due_transition_idx
    ON platform.billing_subscriptions (state, period_end, trial_ends_at, grace_ends_at);


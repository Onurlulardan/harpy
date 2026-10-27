-- +goose Up
CREATE TABLE platform.billing_subscription_changes (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    before_offering_version_id uuid NOT NULL,
    before_price_version_id uuid NOT NULL,
    after_offering_version_id uuid NOT NULL,
    after_price_version_id uuid NOT NULL,
    source text NOT NULL,
    state text NOT NULL,
    effective_at timestamptz NOT NULL,
    source_operation_id uuid NOT NULL,
    decision_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    resolved_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_subscription_changes_pkey PRIMARY KEY (id),
    CONSTRAINT billing_subscription_changes_operation_key
        UNIQUE (tenant_id, source_operation_id),
    CONSTRAINT billing_subscription_changes_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_changes_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_changes_before_offering_fk
        FOREIGN KEY (before_offering_version_id)
        REFERENCES platform.billing_offering_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_changes_after_offering_fk
        FOREIGN KEY (after_offering_version_id)
        REFERENCES platform.billing_offering_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_changes_before_price_fk
        FOREIGN KEY (before_price_version_id)
        REFERENCES platform.billing_price_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_changes_after_price_fk
        FOREIGN KEY (after_price_version_id)
        REFERENCES platform.billing_price_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_changes_source_check CHECK (
        source IN ('manual', 'provider', 'v1_adopted')
    ),
    CONSTRAINT billing_subscription_changes_state_check CHECK (
        (state = 'scheduled' AND resolved_at IS NULL)
        OR (state IN ('applied', 'cancelled') AND resolved_at IS NOT NULL)
    ),
    CONSTRAINT billing_subscription_changes_revision_check CHECK (
        revision > 0 AND effective_at >= created_at
        AND (resolved_at IS NULL OR resolved_at >= created_at)
    )
);

CREATE UNIQUE INDEX billing_subscription_changes_pending_idx
    ON platform.billing_subscription_changes (subscription_id, effective_at)
    WHERE state = 'scheduled';
CREATE INDEX billing_subscription_changes_due_idx
    ON platform.billing_subscription_changes (state, effective_at, subscription_id);

CREATE TABLE platform.billing_subscription_overrides (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    entitlement_definition_id uuid NOT NULL,
    value_type text NOT NULL,
    definition_allows_unlimited boolean NOT NULL,
    boolean_value boolean,
    integer_value bigint,
    unlimited boolean NOT NULL,
    effective_start timestamptz NOT NULL,
    effective_end timestamptz,
    source_operation_id uuid NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_subscription_overrides_pkey PRIMARY KEY (id),
    CONSTRAINT billing_subscription_overrides_start_key UNIQUE (
        subscription_id, entitlement_definition_id, effective_start
    ),
    CONSTRAINT billing_subscription_overrides_operation_key
        UNIQUE (tenant_id, source_operation_id),
    CONSTRAINT billing_subscription_overrides_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_overrides_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_overrides_definition_fk FOREIGN KEY (
        entitlement_definition_id, value_type, definition_allows_unlimited
    ) REFERENCES platform.billing_entitlement_definitions (
        id, value_type, unlimited_allowed
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_overrides_value_check CHECK (
        (value_type = 'boolean' AND boolean_value IS NOT NULL
            AND integer_value IS NULL AND NOT unlimited)
        OR (value_type = 'integer' AND boolean_value IS NULL AND (
            (unlimited AND definition_allows_unlimited AND integer_value IS NULL)
            OR (NOT unlimited AND integer_value >= 0)
        ))
    ),
    CONSTRAINT billing_subscription_overrides_interval_check CHECK (
        effective_end IS NULL OR effective_start < effective_end
    )
);


-- +goose Up
CREATE TABLE platform.billing_meter_periods (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    entitlement_definition_id uuid NOT NULL,
    entitlement_key text COLLATE "C" NOT NULL,
    meter_kind text NOT NULL,
    reset_policy text NOT NULL,
    period_start timestamptz NOT NULL,
    period_end timestamptz NOT NULL,
    state text NOT NULL,
    closed_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_meter_periods_pkey PRIMARY KEY (id),
    CONSTRAINT billing_meter_periods_identity_key UNIQUE (
        tenant_id, entitlement_key, period_start
    ),
    CONSTRAINT billing_meter_periods_parent_key UNIQUE (
        tenant_id, id, entitlement_key, meter_kind
    ),
    CONSTRAINT billing_meter_periods_tenant_id_key_key UNIQUE (
        tenant_id, id, entitlement_key
    ),
    CONSTRAINT billing_meter_periods_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_meter_periods_definition_fk FOREIGN KEY (
        entitlement_definition_id, entitlement_key
    ) REFERENCES platform.billing_entitlement_definitions (id, entitlement_key)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_meter_periods_policy_check CHECK (
        (meter_kind = 'counter'
            AND reset_policy IN ('billing_period', 'calendar_month_utc', 'never'))
        OR (meter_kind = 'gauge' AND reset_policy = 'never')
    ),
    CONSTRAINT billing_meter_periods_interval_check CHECK (
        period_start < period_end
        AND ((state = 'open' AND closed_at IS NULL)
            OR (state = 'closed' AND closed_at >= period_end))
        AND revision > 0
    )
);

CREATE INDEX billing_meter_periods_active_idx
    ON platform.billing_meter_periods (
        tenant_id, entitlement_key, state, period_start, period_end
    );

CREATE TABLE platform.billing_usage_totals (
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    meter_period_id uuid NOT NULL,
    entitlement_key text COLLATE "C" NOT NULL,
    meter_kind text NOT NULL,
    committed_total bigint NOT NULL,
    reserved_total bigint NOT NULL,
    gauge_total bigint NOT NULL,
    revision bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT billing_usage_totals_pkey PRIMARY KEY (meter_period_id),
    CONSTRAINT billing_usage_totals_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_usage_totals_period_fk FOREIGN KEY (
        tenant_id, meter_period_id, entitlement_key, meter_kind
    ) REFERENCES platform.billing_meter_periods (
        tenant_id, id, entitlement_key, meter_kind
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_usage_totals_value_check CHECK (
        committed_total >= 0 AND reserved_total >= 0 AND gauge_total >= 0
        AND revision > 0
        AND ((meter_kind = 'counter' AND gauge_total = 0)
            OR (meter_kind = 'gauge'
                AND committed_total = 0 AND reserved_total = 0))
    )
);

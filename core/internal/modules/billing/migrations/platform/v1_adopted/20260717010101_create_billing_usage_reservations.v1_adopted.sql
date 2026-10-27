-- +goose Up
CREATE TABLE platform.billing_usage_reservations (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    decision_id uuid NOT NULL,
    entitlement_key text COLLATE "C" NOT NULL,
    meter_period_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    amount bigint NOT NULL,
    state text NOT NULL,
    expires_at timestamptz NOT NULL,
    effect_started_at timestamptz,
    effect_receipt_id uuid,
    resolution_receipt_id uuid,
    created_at timestamptz NOT NULL,
    resolved_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_usage_reservations_pkey PRIMARY KEY (id),
    CONSTRAINT billing_usage_reservations_tenant_id_key UNIQUE (tenant_id, id),
    CONSTRAINT billing_usage_reservations_operation_key UNIQUE (
        tenant_id, entitlement_key, source_operation_id
    ),
    CONSTRAINT billing_usage_reservations_decision_key UNIQUE (decision_id),
    CONSTRAINT billing_usage_reservations_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_usage_reservations_period_fk FOREIGN KEY (
        tenant_id, meter_period_id, entitlement_key
    ) REFERENCES platform.billing_meter_periods (
        tenant_id, id, entitlement_key
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_usage_reservations_state_check CHECK (
        (state = 'reserved' AND resolved_at IS NULL)
        OR (state IN ('committed', 'released') AND resolved_at IS NOT NULL)
        OR (state = 'reconcile_required' AND effect_started_at IS NOT NULL
            AND effect_receipt_id IS NOT NULL AND resolved_at IS NULL)
    ),
    CONSTRAINT billing_usage_reservations_release_check CHECK (
        (state = 'reserved' AND resolution_receipt_id IS NULL)
        OR (state = 'committed' AND resolution_receipt_id IS NOT NULL)
        OR (state = 'released' AND (
            effect_started_at IS NULL OR resolution_receipt_id IS NOT NULL
        ))
        OR (state = 'reconcile_required' AND resolution_receipt_id IS NULL)
    ),
    CONSTRAINT billing_usage_reservations_time_check CHECK (
        amount > 0 AND expires_at > created_at AND revision > 0
        AND (effect_started_at IS NULL OR effect_started_at >= created_at)
        AND (resolved_at IS NULL OR resolved_at >= created_at)
    )
);

CREATE INDEX billing_usage_reservations_expiry_idx
    ON platform.billing_usage_reservations (state, expires_at, tenant_id, id)
    WHERE state IN ('reserved', 'reconcile_required');

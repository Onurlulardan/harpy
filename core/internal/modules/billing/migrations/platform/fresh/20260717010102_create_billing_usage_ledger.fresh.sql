-- +goose Up
CREATE TABLE platform.billing_usage_entries (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    decision_id uuid,
    entitlement_key text COLLATE "C" NOT NULL,
    meter_period_id uuid NOT NULL,
    reservation_id uuid,
    entry_kind text NOT NULL,
    amount_delta bigint NOT NULL,
    committed_after bigint NOT NULL,
    reserved_after bigint NOT NULL,
    gauge_after bigint NOT NULL,
    source_operation_id uuid,
    owner_fact_id uuid,
    evidence_sha256 bytea,
    occurred_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_usage_entries_pkey PRIMARY KEY (id),
    CONSTRAINT billing_usage_entries_operation_key UNIQUE (
        tenant_id, entitlement_key, source_operation_id, entry_kind
    ),
    CONSTRAINT billing_usage_entries_fact_key UNIQUE (
        tenant_id, entitlement_key, owner_fact_id
    ),
    CONSTRAINT billing_usage_entries_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_usage_entries_period_fk FOREIGN KEY (
        tenant_id, meter_period_id, entitlement_key
    ) REFERENCES platform.billing_meter_periods (
        tenant_id, id, entitlement_key
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_usage_entries_reservation_fk FOREIGN KEY (
        tenant_id, reservation_id
    ) REFERENCES platform.billing_usage_reservations (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_usage_entries_kind_check CHECK (
        entry_kind IN ('opening', 'reserve', 'commit', 'release', 'gauge_delta')
        AND amount_delta <> 0
        AND committed_after >= 0 AND reserved_after >= 0 AND gauge_after >= 0
    ),
    CONSTRAINT billing_usage_entries_identity_check CHECK (
        (entry_kind = 'opening' AND reservation_id IS NULL
            AND decision_id IS NULL
            AND source_operation_id IS NULL AND owner_fact_id IS NULL
            AND pg_catalog.octet_length(evidence_sha256) = 32)
        OR (entry_kind IN ('reserve', 'commit', 'release')
            AND reservation_id IS NOT NULL AND decision_id IS NOT NULL
            AND source_operation_id IS NOT NULL
            AND owner_fact_id IS NULL AND evidence_sha256 IS NULL)
        OR (entry_kind = 'gauge_delta' AND reservation_id IS NULL
            AND decision_id IS NOT NULL
            AND source_operation_id IS NULL AND owner_fact_id IS NOT NULL
            AND evidence_sha256 IS NULL)
    ),
    CONSTRAINT billing_usage_entries_time_check CHECK (created_at >= occurred_at)
);

CREATE INDEX billing_usage_entries_period_time_idx
    ON platform.billing_usage_entries (meter_period_id, occurred_at, id);

CREATE TABLE platform.billing_entitlement_decisions (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    subscription_id uuid,
    entitlement_definition_id uuid NOT NULL,
    entitlement_key text COLLATE "C" NOT NULL,
    meter_period_id uuid,
    source_operation_id uuid NOT NULL,
    requested_amount bigint NOT NULL,
    current_amount bigint NOT NULL,
    limit_amount bigint,
    unlimited boolean NOT NULL,
    allowed boolean NOT NULL,
    soft_exceeded boolean NOT NULL,
    reason text COLLATE "C" NOT NULL,
    correlation_id text COLLATE "C" NOT NULL,
    decided_at timestamptz NOT NULL,
    retain_until timestamptz NOT NULL,
    CONSTRAINT billing_entitlement_decisions_pkey PRIMARY KEY (id),
    CONSTRAINT billing_entitlement_decisions_operation_key UNIQUE (
        tenant_id, entitlement_key, source_operation_id
    ),
    CONSTRAINT billing_entitlement_decisions_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_entitlement_decisions_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_entitlement_decisions_definition_fk FOREIGN KEY (
        entitlement_definition_id, entitlement_key
    ) REFERENCES platform.billing_entitlement_definitions (id, entitlement_key)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_entitlement_decisions_period_fk FOREIGN KEY (
        tenant_id, meter_period_id, entitlement_key
    ) REFERENCES platform.billing_meter_periods (tenant_id, id, entitlement_key)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_entitlement_decisions_amount_check CHECK (
        requested_amount <> 0 AND current_amount >= 0
        AND ((unlimited AND limit_amount IS NULL)
            OR (NOT unlimited AND limit_amount >= 0))
    ),
    CONSTRAINT billing_entitlement_decisions_reason_check CHECK (
        reason IN (
            'allowed', 'base_grant', 'subscription_grant', 'override_grant',
            'missing_grant', 'hard_limit', 'soft_limit', 'dependency_not_ready'
        )
        AND correlation_id ~ '^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$'
        AND retain_until > decided_at
    )
);

ALTER TABLE platform.billing_usage_reservations
    ADD CONSTRAINT billing_usage_reservations_decision_fk
    FOREIGN KEY (decision_id)
    REFERENCES platform.billing_entitlement_decisions (id)
    ON UPDATE NO ACTION ON DELETE RESTRICT;

ALTER TABLE platform.billing_usage_entries
    ADD CONSTRAINT billing_usage_entries_decision_fk
    FOREIGN KEY (decision_id)
    REFERENCES platform.billing_entitlement_decisions (id)
    ON UPDATE NO ACTION ON DELETE RESTRICT;

CREATE INDEX billing_entitlement_decisions_tenant_time_idx
    ON platform.billing_entitlement_decisions (tenant_id, decided_at, id);

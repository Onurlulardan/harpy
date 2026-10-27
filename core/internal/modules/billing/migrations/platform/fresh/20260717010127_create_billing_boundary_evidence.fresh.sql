-- +goose Up
ALTER TABLE platform.billing_subscription_boundary_jobs
    ADD CONSTRAINT billing_subscription_boundary_jobs_tenant_id_key
        UNIQUE (tenant_id, id);
ALTER TABLE platform.billing_provider_events
    ADD CONSTRAINT billing_provider_events_tenant_id_key
        UNIQUE (tenant_id, id);

CREATE TABLE platform.billing_subscription_boundary_evidence (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    boundary_job_id uuid NOT NULL,
    provider_event_id uuid NOT NULL,
    boundary_kind text NOT NULL,
    due_at timestamptz NOT NULL,
    expected_revision bigint NOT NULL,
    outcome text NOT NULL,
    next_period_start timestamptz,
    next_period_end timestamptz,
    grace_ends_at timestamptz,
    recorded_at timestamptz NOT NULL,
    CONSTRAINT billing_subscription_boundary_evidence_pkey PRIMARY KEY (id),
    CONSTRAINT billing_subscription_boundary_evidence_event_key
        UNIQUE (tenant_id, provider_event_id),
    CONSTRAINT billing_subscription_boundary_evidence_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_boundary_evidence_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_boundary_evidence_job_fk FOREIGN KEY (
        tenant_id, boundary_job_id
    ) REFERENCES platform.billing_subscription_boundary_jobs (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_boundary_evidence_provider_fk FOREIGN KEY (
        tenant_id, provider_event_id
    ) REFERENCES platform.billing_provider_events (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_boundary_evidence_boundary_check CHECK (
        boundary_kind IN ('manual_change', 'trial_end', 'period_end')
        AND expected_revision > 0
    ),
    CONSTRAINT billing_subscription_boundary_evidence_outcome_check CHECK (
        (outcome = 'paid'
            AND next_period_start = due_at
            AND next_period_end > next_period_start
            AND grace_ends_at IS NULL)
        OR (outcome = 'grace'
            AND boundary_kind IN ('manual_change', 'trial_end')
            AND next_period_start IS NULL AND next_period_end IS NULL
            AND grace_ends_at > due_at)
        OR (outcome = 'unpaid'
            AND next_period_start IS NULL AND next_period_end IS NULL
            AND grace_ends_at IS NULL)
    )
);

CREATE INDEX billing_subscription_boundary_evidence_lookup_idx
    ON platform.billing_subscription_boundary_evidence (
        tenant_id, boundary_job_id, expected_revision, recorded_at, id
    );

CREATE TRIGGER billing_subscription_boundary_evidence_authority_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_subscription_boundary_evidence
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_subscription_boundary_evidence
    ENABLE ALWAYS TRIGGER billing_subscription_boundary_evidence_authority_guard;
CREATE TRIGGER billing_subscription_boundary_evidence_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE
ON platform.billing_subscription_boundary_evidence
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_subscription_boundary_evidence
    ENABLE ALWAYS TRIGGER billing_subscription_boundary_evidence_immutable_guard;

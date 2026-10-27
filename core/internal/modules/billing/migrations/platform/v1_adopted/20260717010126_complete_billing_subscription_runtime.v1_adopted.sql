-- +goose Up
ALTER TABLE platform.billing_subscription_operations
    ADD COLUMN result_provider_account_id uuid NOT NULL,
    ADD COLUMN result_offering_version_id uuid NOT NULL,
    ADD COLUMN result_price_version_id uuid NOT NULL,
    ADD COLUMN result_currency character(3) COLLATE "C" NOT NULL,
    ADD COLUMN result_period_start timestamptz,
    ADD COLUMN result_period_end timestamptz,
    ADD COLUMN result_trial_ends_at timestamptz,
    ADD COLUMN result_grace_ends_at timestamptz,
    ADD COLUMN result_cancel_at_period_end boolean NOT NULL,
    ADD COLUMN result_cancel_at timestamptz,
    ADD COLUMN result_updated_at timestamptz NOT NULL,
    ADD CONSTRAINT billing_subscription_operations_result_provider_fk
        FOREIGN KEY (result_provider_account_id)
        REFERENCES platform.billing_provider_accounts (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    ADD CONSTRAINT billing_subscription_operations_result_offering_fk
        FOREIGN KEY (result_offering_version_id)
        REFERENCES platform.billing_offering_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    ADD CONSTRAINT billing_subscription_operations_result_price_fk
        FOREIGN KEY (result_price_version_id)
        REFERENCES platform.billing_price_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    ADD CONSTRAINT billing_subscription_operations_snapshot_check CHECK (
        (result_period_start IS NULL AND result_period_end IS NULL)
        OR (result_period_start IS NOT NULL AND result_period_end IS NOT NULL
            AND result_period_start < result_period_end)
    ),
    ADD CONSTRAINT billing_subscription_operations_window_check CHECK (
        (result_state <> 'trialing' OR (
            result_trial_ends_at IS NOT NULL
            AND result_trial_ends_at = result_period_end
        ))
        AND (result_state <> 'past_due' OR result_grace_ends_at IS NOT NULL)
        AND result_cancel_at_period_end = (result_cancel_at IS NOT NULL)
        AND (result_cancel_at IS NULL OR result_cancel_at = result_period_end)
        AND result_updated_at <= completed_at
    );

CREATE TABLE platform.billing_subscription_audit (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    source_kind text NOT NULL,
    kind text COLLATE "C" NOT NULL,
    actor_id uuid,
    before_revision bigint NOT NULL,
    after_revision bigint NOT NULL,
    occurred_at timestamptz NOT NULL,
    reason text NOT NULL,
    CONSTRAINT billing_subscription_audit_pkey PRIMARY KEY (id),
    CONSTRAINT billing_subscription_audit_operation_key
        UNIQUE (tenant_id, source_operation_id),
    CONSTRAINT billing_subscription_audit_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_audit_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_audit_operation_fk FOREIGN KEY (
        tenant_id, source_operation_id
    ) REFERENCES platform.billing_subscription_operations (
        tenant_id, source_operation_id
    ) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_audit_shape_check CHECK (
        source_kind IN ('actor', 'provider', 'job')
        AND kind ~ '^[a-z][a-z0-9._]{2,127}$'
        AND ((source_kind = 'actor' AND actor_id IS NOT NULL)
            OR (source_kind <> 'actor' AND actor_id IS NULL))
        AND before_revision > 0 AND after_revision = before_revision + 1
        AND pg_catalog.octet_length(reason) <= 512
    )
);

CREATE TABLE platform.billing_subscription_boundary_jobs (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    effect_id uuid NOT NULL,
    boundary_kind text NOT NULL,
    run_at timestamptz NOT NULL,
    expected_revision bigint NOT NULL,
    payload_revision integer NOT NULL,
    payload_reference text COLLATE "C" NOT NULL,
    payload_sha256 bytea NOT NULL,
    created_at timestamptz NOT NULL,
    CONSTRAINT billing_subscription_boundary_jobs_pkey PRIMARY KEY (id),
    CONSTRAINT billing_subscription_boundary_jobs_effect_key
        UNIQUE (tenant_id, effect_id),
    CONSTRAINT billing_subscription_boundary_jobs_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_boundary_jobs_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_boundary_jobs_async_fk FOREIGN KEY (id)
        REFERENCES platform.jobs (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
        DEFERRABLE INITIALLY DEFERRED,
    CONSTRAINT billing_subscription_boundary_jobs_shape_check CHECK (
        boundary_kind IN (
            'cancellation', 'manual_change', 'trial_end', 'grace_end', 'period_end'
        )
        AND expected_revision > 0 AND payload_revision = 1
        AND pg_catalog.octet_length(payload_reference) BETWEEN 1 AND 2048
        AND pg_catalog.octet_length(payload_sha256) = 32
    )
);

CREATE INDEX billing_subscription_boundary_jobs_lookup_idx
    ON platform.billing_subscription_boundary_jobs (
        tenant_id, subscription_id, boundary_kind, expected_revision
    );

CREATE TRIGGER billing_subscription_audit_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_audit
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_subscription_audit
    ENABLE ALWAYS TRIGGER billing_subscription_audit_authority_guard;
CREATE TRIGGER billing_subscription_audit_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_audit
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_subscription_audit
    ENABLE ALWAYS TRIGGER billing_subscription_audit_immutable_guard;

CREATE TRIGGER billing_subscription_boundary_jobs_authority_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_boundary_jobs
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_tenant_row_v1();
ALTER TABLE platform.billing_subscription_boundary_jobs
    ENABLE ALWAYS TRIGGER billing_subscription_boundary_jobs_authority_guard;
CREATE TRIGGER billing_subscription_boundary_jobs_immutable_guard
BEFORE INSERT OR UPDATE OR DELETE ON platform.billing_subscription_boundary_jobs
FOR EACH ROW EXECUTE FUNCTION platform.billing_guard_immutable_v1();
ALTER TABLE platform.billing_subscription_boundary_jobs
    ENABLE ALWAYS TRIGGER billing_subscription_boundary_jobs_immutable_guard;

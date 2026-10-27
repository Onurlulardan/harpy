-- +goose Up
CREATE TABLE platform.billing_trial_claims (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    offering_version_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    claimed_at timestamptz NOT NULL,
    CONSTRAINT billing_trial_claims_pkey PRIMARY KEY (id),
    CONSTRAINT billing_trial_claims_policy_key
        UNIQUE (tenant_id, offering_version_id),
    CONSTRAINT billing_trial_claims_operation_key
        UNIQUE (tenant_id, source_operation_id),
    CONSTRAINT billing_trial_claims_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_trial_claims_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_trial_claims_offering_fk FOREIGN KEY (offering_version_id)
        REFERENCES platform.billing_offering_versions (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT
);


-- +goose Up
CREATE TABLE platform.billing_subscription_operations (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    subscription_id uuid NOT NULL,
    source_operation_id uuid NOT NULL,
    source_kind text NOT NULL,
    operation_kind text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    result_sha256 bytea NOT NULL,
    expected_revision bigint NOT NULL,
    result_revision bigint NOT NULL,
    result_state text NOT NULL,
    completed_at timestamptz NOT NULL,
    CONSTRAINT billing_subscription_operations_pkey PRIMARY KEY (id),
    CONSTRAINT billing_subscription_operations_source_key
        UNIQUE (tenant_id, source_operation_id),
    CONSTRAINT billing_subscription_operations_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_operations_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_subscription_operations_source_check CHECK (
        source_kind IN ('actor', 'provider', 'job', 'v1_adopted')
        AND operation_kind ~ '^[a-z][a-z0-9._]{2,127}$'
    ),
    CONSTRAINT billing_subscription_operations_result_check CHECK (
        pg_catalog.octet_length(request_sha256) = 32
        AND pg_catalog.octet_length(result_sha256) = 32
        AND expected_revision >= 0
        AND result_revision = expected_revision + 1
        AND result_state IN (
            'pending', 'trialing', 'active', 'past_due',
            'paused', 'cancelled', 'expired'
        )
    )
);

CREATE INDEX billing_subscription_operations_subscription_idx
    ON platform.billing_subscription_operations (
        tenant_id, subscription_id, completed_at, id
    );

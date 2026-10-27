-- +goose Up
CREATE TABLE platform.billing_provider_events (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    webhook_receipt_id uuid NOT NULL,
    provider_object_mapping_id uuid NOT NULL,
    normalized_schema text COLLATE "C" NOT NULL,
    normalized_version integer NOT NULL,
    ordering_mode text NOT NULL,
    ordering_sequence bigint NOT NULL,
    ordering_previous bigint NOT NULL,
    canonical_fact bytea NOT NULL,
    occurred_at timestamptz NOT NULL,
    normalized_at timestamptz NOT NULL,
    normalized_sha256 bytea NOT NULL,
    CONSTRAINT billing_provider_events_pkey PRIMARY KEY (id),
    CONSTRAINT billing_provider_events_receipt_key UNIQUE (webhook_receipt_id),
    CONSTRAINT billing_provider_events_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_events_receipt_fk FOREIGN KEY (
        tenant_id, webhook_receipt_id
    ) REFERENCES platform.billing_webhook_receipts (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_events_mapping_fk FOREIGN KEY (
        tenant_id, provider_object_mapping_id
    ) REFERENCES platform.billing_provider_objects (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_events_schema_check CHECK (
        normalized_schema ~ '^[a-z][a-z0-9_.]{2,127}$'
        AND normalized_version > 0
    ),
    CONSTRAINT billing_provider_events_ordering_check CHECK (
        (ordering_mode = 'contiguous' AND ordering_sequence > 0
            AND ordering_previous = ordering_sequence - 1)
        OR (ordering_mode = 'requires_reconciliation'
            AND ordering_sequence = 0 AND ordering_previous = 0)
    ),
    CONSTRAINT billing_provider_events_integrity_check CHECK (
        normalized_at >= occurred_at
        AND pg_catalog.octet_length(canonical_fact) BETWEEN 1 AND 65536
        AND pg_catalog.octet_length(normalized_sha256) = 32
        AND normalized_sha256 = pg_catalog.sha256(canonical_fact)
    )
);

CREATE TABLE platform.billing_provider_commands (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    provider_account_id uuid NOT NULL,
    provider_object_mapping_id uuid,
    subscription_id uuid,
    command_kind text NOT NULL,
    effect_key text COLLATE "C" NOT NULL,
    request_sha256 bytea NOT NULL,
    state text NOT NULL,
    provider_receipt_reference text COLLATE "C",
    last_error_code text COLLATE "C",
    attempts integer NOT NULL,
    requested_at timestamptz NOT NULL,
    completed_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_provider_commands_pkey PRIMARY KEY (id),
    CONSTRAINT billing_provider_commands_effect_key UNIQUE (
        provider_account_id, effect_key
    ),
    CONSTRAINT billing_provider_commands_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_commands_provider_fk FOREIGN KEY (provider_account_id)
        REFERENCES platform.billing_provider_accounts (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_commands_mapping_fk FOREIGN KEY (
        tenant_id, provider_object_mapping_id
    ) REFERENCES platform.billing_provider_objects (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_commands_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_commands_kind_check CHECK (
        command_kind IN ('checkout', 'change', 'cancel', 'refund', 'reconcile')
        AND pg_catalog.octet_length(effect_key) BETWEEN 1 AND 256
        AND pg_catalog.octet_length(request_sha256) = 32
    ),
    CONSTRAINT billing_provider_commands_state_check CHECK (
        (state IN ('queued', 'running') AND completed_at IS NULL)
        OR (state IN ('succeeded', 'failed', 'reconcile_required')
            AND completed_at IS NOT NULL)
    ),
    CONSTRAINT billing_provider_commands_receipt_check CHECK (
        (state = 'succeeded' AND provider_receipt_reference IS NOT NULL
            AND last_error_code IS NULL)
        OR (state <> 'succeeded')
    ),
    CONSTRAINT billing_provider_commands_attempt_check CHECK (
        attempts BETWEEN 0 AND 25 AND revision > 0
        AND (completed_at IS NULL OR completed_at >= requested_at)
    )
);

CREATE INDEX billing_provider_commands_claim_idx
    ON platform.billing_provider_commands (state, requested_at, id)
    WHERE state IN ('queued', 'running');

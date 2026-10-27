-- +goose Up
CREATE TABLE platform.billing_provider_objects (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    provider_account_id uuid NOT NULL,
    object_kind text NOT NULL,
    provider_object_id text COLLATE "C" NOT NULL,
    subscription_id uuid,
    invoice_id uuid,
    created_at timestamptz NOT NULL,
    retired_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_provider_objects_pkey PRIMARY KEY (id),
    CONSTRAINT billing_provider_objects_tenant_id_key UNIQUE (tenant_id, id),
    CONSTRAINT billing_provider_objects_tenant_provider_key
        UNIQUE (tenant_id, id, provider_account_id),
    CONSTRAINT billing_provider_objects_identity_key UNIQUE (
        provider_account_id, object_kind, provider_object_id
    ),
    CONSTRAINT billing_provider_objects_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_objects_provider_fk FOREIGN KEY (provider_account_id)
        REFERENCES platform.billing_provider_accounts (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_objects_subscription_fk FOREIGN KEY (
        tenant_id, subscription_id
    ) REFERENCES platform.billing_subscriptions (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_objects_invoice_fk FOREIGN KEY (
        tenant_id, invoice_id
    ) REFERENCES platform.billing_invoices (tenant_id, id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_objects_kind_check CHECK (
        (object_kind = 'customer' AND subscription_id IS NULL AND invoice_id IS NULL)
        OR (object_kind = 'subscription' AND subscription_id IS NOT NULL)
        OR (object_kind IN ('invoice', 'payment') AND invoice_id IS NOT NULL)
    ),
    CONSTRAINT billing_provider_objects_id_check CHECK (
        pg_catalog.octet_length(provider_object_id) BETWEEN 1 AND 512
        AND revision > 0
        AND (retired_at IS NULL OR retired_at >= created_at)
    )
);

CREATE INDEX billing_provider_objects_tenant_lookup_idx
    ON platform.billing_provider_objects (
        tenant_id, object_kind, subscription_id, invoice_id
    );

CREATE TABLE platform.billing_webhook_receipts (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    provider_account_id uuid NOT NULL,
    provider_object_mapping_id uuid NOT NULL,
    provider_event_id text COLLATE "C" NOT NULL,
    body_sha256 bytea NOT NULL,
    authentication_digest bytea NOT NULL,
    verification_revision integer NOT NULL,
    event_occurred_at timestamptz NOT NULL,
    signature_issued_at timestamptz NOT NULL,
    algorithm text COLLATE "C",
    nonce bytea,
    encrypted_body bytea,
    wrapped_dek bytea,
    key_id text COLLATE "C" NOT NULL,
    verified_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL,
    processor_state text NOT NULL,
    processed_at timestamptz,
    raw_body_expires_at timestamptz NOT NULL,
    retain_until timestamptz NOT NULL,
    revision bigint NOT NULL,
    CONSTRAINT billing_webhook_receipts_pkey PRIMARY KEY (id),
    CONSTRAINT billing_webhook_receipts_event_key
        UNIQUE (provider_account_id, provider_event_id),
    CONSTRAINT billing_webhook_receipts_tenant_id_key UNIQUE (tenant_id, id),
    CONSTRAINT billing_webhook_receipts_authority_fk FOREIGN KEY (
        authority_id, tenant_id
    ) REFERENCES platform.billing_tenant_authorities (authority_id, tenant_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_webhook_receipts_provider_fk FOREIGN KEY (provider_account_id)
        REFERENCES platform.billing_provider_accounts (id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_webhook_receipts_mapping_fk FOREIGN KEY (
        tenant_id, provider_object_mapping_id, provider_account_id
    ) REFERENCES platform.billing_provider_objects (
        tenant_id, id, provider_account_id
    )
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_webhook_receipts_crypto_check CHECK (
        pg_catalog.octet_length(body_sha256) = 32
        AND pg_catalog.octet_length(authentication_digest) = 32
        AND verification_revision > 0
        AND pg_catalog.octet_length(key_id) BETWEEN 1 AND 256
        AND ((encrypted_body IS NULL AND wrapped_dek IS NULL
                AND algorithm IS NULL AND nonce IS NULL)
            OR (algorithm = 'AES-256-GCM'
                AND pg_catalog.octet_length(nonce) = 12
                AND pg_catalog.octet_length(encrypted_body) BETWEEN 17 AND 1048592
                AND pg_catalog.octet_length(wrapped_dek) BETWEEN 1 AND 8192))
    ),
    CONSTRAINT billing_webhook_receipts_state_check CHECK (
        processor_state IN (
            'received', 'processing', 'applied', 'reconcile_required', 'rejected'
        )
        AND ((processor_state IN ('received', 'processing') AND processed_at IS NULL)
            OR (processor_state IN ('applied', 'reconcile_required', 'rejected')
                AND processed_at IS NOT NULL))
    ),
    CONSTRAINT billing_webhook_receipts_time_check CHECK (
        event_occurred_at <= received_at AND signature_issued_at <= received_at
        AND verified_at <= received_at AND raw_body_expires_at > received_at
        AND retain_until >= raw_body_expires_at AND revision > 0
    )
);

CREATE INDEX billing_webhook_receipts_processor_idx
    ON platform.billing_webhook_receipts (processor_state, received_at, id);
CREATE INDEX billing_webhook_receipts_retention_idx
    ON platform.billing_webhook_receipts (raw_body_expires_at, retain_until);

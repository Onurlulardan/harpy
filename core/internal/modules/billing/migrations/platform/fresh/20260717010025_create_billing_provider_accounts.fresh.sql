-- +goose Up
CREATE TABLE platform.billing_provider_accounts (
    id uuid NOT NULL,
    authority_id uuid NOT NULL,
    provider_key text COLLATE "C" NOT NULL,
    provider_account_key text COLLATE "C" NOT NULL,
    lifecycle text NOT NULL,
    supports_checkout boolean NOT NULL,
    supports_webhooks boolean NOT NULL,
    supports_proration boolean NOT NULL,
    supports_refunds boolean NOT NULL,
    created_at timestamptz NOT NULL,
    disabled_at timestamptz,
    revision bigint NOT NULL,
    CONSTRAINT billing_provider_accounts_pkey PRIMARY KEY (id),
    CONSTRAINT billing_provider_accounts_identity_key
        UNIQUE (provider_key, provider_account_key),
    CONSTRAINT billing_provider_accounts_authority_fk FOREIGN KEY (authority_id)
        REFERENCES platform.billing_global_authorities (authority_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT,
    CONSTRAINT billing_provider_accounts_key_check CHECK (
        provider_key ~ '^[a-z][a-z0-9._-]{0,62}$'
        AND provider_account_key ~ '^[a-zA-Z0-9][a-zA-Z0-9._:-]{0,254}$'
    ),
    CONSTRAINT billing_provider_accounts_lifecycle_check CHECK (
        (lifecycle = 'active' AND disabled_at IS NULL)
        OR (lifecycle = 'disabled' AND disabled_at IS NOT NULL)
    ),
    CONSTRAINT billing_provider_accounts_manual_check CHECK (
        provider_key <> 'manual'
        OR (NOT supports_checkout AND NOT supports_webhooks
            AND NOT supports_proration AND NOT supports_refunds)
    ),
    CONSTRAINT billing_provider_accounts_revision_check CHECK (revision > 0),
    CONSTRAINT billing_provider_accounts_time_check CHECK (
        disabled_at IS NULL OR disabled_at >= created_at
    )
);


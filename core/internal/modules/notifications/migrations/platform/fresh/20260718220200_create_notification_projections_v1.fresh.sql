-- +goose Up
CREATE TABLE platform.notification_realtime_tickets (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    token_hash bytea NOT NULL CHECK (pg_catalog.octet_length(token_hash) = 32),
    authorization_version bigint NOT NULL CHECK (authorization_version > 0),
    origin text COLLATE "C" NOT NULL
        CHECK (origin ~ '^https://[A-Za-z0-9.-]+(:[1-9][0-9]{0,4})?$'),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    expires_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(expires_at)),
    consumed_at timestamptz,
    CONSTRAINT notification_realtime_tickets_token_key UNIQUE (token_hash),
    CHECK (expires_at = created_at + interval '60 seconds'),
    CHECK (consumed_at IS NULL OR (
        pg_catalog.isfinite(consumed_at) AND consumed_at >= created_at
        AND consumed_at < expires_at
    ))
);

CREATE INDEX notification_realtime_tickets_purge_idx
    ON platform.notification_realtime_tickets (
        COALESCE(consumed_at, expires_at), id
    );

CREATE TABLE platform.notification_provider_locators (
    provider_kind text COLLATE "C" NOT NULL CHECK (provider_kind IN ('smtp','telegram')),
    provider_account_digest bytea NOT NULL
        CHECK (pg_catalog.octet_length(provider_account_digest) = 32),
    external_message_digest bytea NOT NULL
        CHECK (pg_catalog.octet_length(external_message_digest) = 32),
    tenant_id uuid NOT NULL,
    delivery_id uuid NOT NULL,
    accepted_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(accepted_at)),
    expires_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(expires_at)),
    PRIMARY KEY (provider_kind, provider_account_digest, external_message_digest),
    CHECK (expires_at = accepted_at + interval '90 days')
);

CREATE INDEX notification_provider_locators_purge_idx
    ON platform.notification_provider_locators (expires_at, delivery_id);

CREATE TABLE platform.notification_legacy_connection_hits (
    id uuid PRIMARY KEY,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    consumer_manifest_digest bytea NOT NULL
        CHECK (pg_catalog.octet_length(consumer_manifest_digest) = 32),
    occurred_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(occurred_at))
);

CREATE INDEX notification_legacy_connection_hits_retirement_idx
    ON platform.notification_legacy_connection_hits (occurred_at DESC, id);

GRANT SELECT ON TABLE platform.notification_realtime_tickets TO harpy_runtime;
GRANT INSERT (id, tenant_id, user_id, token_hash, authorization_version, origin, created_at, expires_at, consumed_at) ON platform.notification_realtime_tickets TO harpy_runtime;
GRANT UPDATE (consumed_at) ON platform.notification_realtime_tickets TO harpy_runtime;
GRANT SELECT ON TABLE platform.notification_provider_locators TO harpy_runtime;
GRANT INSERT (provider_kind, provider_account_digest, external_message_digest, tenant_id, delivery_id, accepted_at, expires_at) ON platform.notification_provider_locators TO harpy_runtime;
GRANT SELECT ON TABLE platform.notification_legacy_connection_hits TO harpy_runtime;
GRANT INSERT (id, tenant_id, user_id, consumer_manifest_digest, occurred_at) ON platform.notification_legacy_connection_hits TO harpy_runtime;

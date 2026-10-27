-- +goose Up
CREATE TABLE notifications_deliveries (
    id uuid PRIMARY KEY,
    message_id uuid NOT NULL,
    provider_config_version_id uuid,
    channel text COLLATE "C" NOT NULL CHECK (channel IN ('email','telegram')),
    state text COLLATE "C" NOT NULL CHECK (state IN (
        'pending','suppressed','sending','accepted','delivered','bounced','failed','cancelled'
    )),
    attempt integer NOT NULL CHECK (attempt BETWEEN 0 AND 8),
    maximum_attempts integer NOT NULL CHECK (maximum_attempts = 8),
    lease_owner text COLLATE "C" NOT NULL DEFAULT '',
    lease_generation bigint NOT NULL DEFAULT 0 CHECK (lease_generation >= 0),
    lease_until timestamptz,
    not_before timestamptz NOT NULL CHECK (pg_catalog.isfinite(not_before)),
    outcome_code text COLLATE "C" NOT NULL DEFAULT ''
        CHECK (outcome_code = '' OR outcome_code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    outcome_unknown boolean NOT NULL DEFAULT false,
    mandatory boolean NOT NULL,
    row_version bigint NOT NULL CHECK (row_version > 0),
    CONSTRAINT notifications_deliveries_message_channel_key UNIQUE (message_id, channel),
    CONSTRAINT notifications_deliveries_message_fk FOREIGN KEY (message_id)
        REFERENCES notifications_messages(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (
        (state = 'sending' AND lease_owner <> '' AND lease_generation > 0
            AND lease_until IS NOT NULL AND pg_catalog.isfinite(lease_until))
        OR (state <> 'sending' AND lease_owner = '' AND lease_until IS NULL)
    ),
    CHECK (NOT outcome_unknown OR state = 'failed')
);

CREATE INDEX notifications_deliveries_claim_idx
    ON notifications_deliveries (not_before, id) WHERE state = 'pending';

CREATE TABLE notifications_delivery_attempts (
    id uuid PRIMARY KEY,
    delivery_id uuid NOT NULL,
    attempt_number integer NOT NULL CHECK (attempt_number BETWEEN 1 AND 8),
    lease_generation bigint NOT NULL CHECK (lease_generation > 0),
    provider_kind text COLLATE "C" NOT NULL CHECK (provider_kind IN ('smtp','telegram')),
    response_class text COLLATE "C" NOT NULL
        CHECK (pg_catalog.octet_length(response_class) BETWEEN 1 AND 64),
    provider_reference_digest bytea NOT NULL
        CHECK (pg_catalog.octet_length(provider_reference_digest) = 32),
    status_code text COLLATE "C" NOT NULL
        CHECK (pg_catalog.octet_length(status_code) <= 32),
    error_code text COLLATE "C" NOT NULL
        CHECK (error_code ~ '^[A-Z][A-Z0-9_]{1,63}$'),
    started_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(started_at)),
    finished_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(finished_at)),
    CONSTRAINT notifications_delivery_attempts_delivery_attempt_key
        UNIQUE (delivery_id, attempt_number),
    CONSTRAINT notifications_delivery_attempts_delivery_fk FOREIGN KEY (delivery_id)
        REFERENCES notifications_deliveries(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (finished_at >= started_at)
);

CREATE TABLE notifications_delivery_receipts (
    id uuid PRIMARY KEY,
    delivery_id uuid NOT NULL,
    provider_reference_digest bytea NOT NULL
        CHECK (pg_catalog.octet_length(provider_reference_digest) = 32),
    outcome text COLLATE "C" NOT NULL CHECK (outcome IN ('delivered','bounced','complained')),
    signature_key_id text COLLATE "C" NOT NULL
        CHECK (pg_catalog.octet_length(signature_key_id) BETWEEN 1 AND 256),
    received_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(received_at)),
    CONSTRAINT notifications_delivery_receipts_effect_key
        UNIQUE (delivery_id, provider_reference_digest, outcome),
    CONSTRAINT notifications_delivery_receipts_delivery_fk FOREIGN KEY (delivery_id)
        REFERENCES notifications_deliveries(id) ON UPDATE NO ACTION ON DELETE RESTRICT
);

CREATE TABLE notifications_suppressed_targets (
    channel text COLLATE "C" NOT NULL CHECK (channel IN ('email','telegram')),
    target_digest bytea NOT NULL CHECK (pg_catalog.octet_length(target_digest) = 32),
    reason text COLLATE "C" NOT NULL CHECK (reason IN ('bounce','complaint','operator')),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    PRIMARY KEY (channel, target_digest)
);

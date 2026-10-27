-- +goose Up
CREATE TABLE notifications_inbox_counters (
    recipient_id uuid PRIMARY KEY,
    current_revision bigint NOT NULL CHECK (current_revision >= 0)
);

CREATE TABLE notifications_messages (
    id uuid PRIMARY KEY,
    recipient_id uuid NOT NULL,
    template_version_id uuid NOT NULL,
    producer text COLLATE "C" NOT NULL CHECK (pg_catalog.octet_length(producer) BETWEEN 1 AND 64),
    notification_type text COLLATE "C" NOT NULL
        CHECK (pg_catalog.octet_length(notification_type) BETWEEN 1 AND 96),
    category text COLLATE "C" NOT NULL CHECK (category IN (
        'security','workflow_action','billing','document','system','product'
    )),
    template_kind text COLLATE "C" NOT NULL CHECK (template_kind IN ('system_html','plain_text')),
    locale text COLLATE "C" NOT NULL CHECK (pg_catalog.octet_length(locale) BETWEEN 2 AND 16),
    subject text NOT NULL,
    body text NOT NULL,
    data_digest bytea NOT NULL CHECK (pg_catalog.octet_length(data_digest) = 32),
    subject_digest bytea NOT NULL CHECK (pg_catalog.octet_length(subject_digest) = 32),
    body_digest bytea NOT NULL CHECK (pg_catalog.octet_length(body_digest) = 32),
    row_version bigint NOT NULL CHECK (row_version > 0),
    inbox_revision bigint NOT NULL CHECK (inbox_revision > 0),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    read_at timestamptz,
    dismissed_at timestamptz,
    purged_at timestamptz,
    CONSTRAINT notifications_messages_recipient_revision_key
        UNIQUE (recipient_id, inbox_revision),
    CHECK (
        (purged_at IS NULL AND pg_catalog.octet_length(subject) BETWEEN 1 AND 2048
            AND pg_catalog.octet_length(body) BETWEEN 1 AND 65536)
        OR (purged_at IS NOT NULL AND subject = '' AND body = '')
    ),
    CHECK (read_at IS NULL OR pg_catalog.isfinite(read_at)),
    CHECK (dismissed_at IS NULL OR pg_catalog.isfinite(dismissed_at)),
    CHECK (purged_at IS NULL OR pg_catalog.isfinite(purged_at))
);

CREATE INDEX notifications_messages_inbox_idx
    ON notifications_messages (recipient_id, inbox_revision DESC, id)
    WHERE dismissed_at IS NULL;

CREATE TABLE notifications_effect_receipts (
    producer text COLLATE "C" NOT NULL,
    effect_key_hash bytea NOT NULL CHECK (pg_catalog.octet_length(effect_key_hash) = 32),
    recipient_id uuid NOT NULL,
    canonical_digest bytea NOT NULL CHECK (pg_catalog.octet_length(canonical_digest) = 32),
    message_id uuid NOT NULL,
    legacy_receipt_reference text COLLATE "C" NOT NULL DEFAULT '',
    outcome text COLLATE "C" NOT NULL CHECK (outcome IN (
        'pending','created','suppressed','outcome_unknown'
    )),
    created_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(created_at)),
    retain_until timestamptz NOT NULL CHECK (pg_catalog.isfinite(retain_until)),
    PRIMARY KEY (producer, effect_key_hash, recipient_id),
    CONSTRAINT notifications_effect_receipts_message_fk FOREIGN KEY (message_id)
        REFERENCES notifications_messages(id) ON UPDATE NO ACTION ON DELETE RESTRICT,
    CHECK (retain_until >= created_at + interval '365 days'),
    CHECK (pg_catalog.octet_length(legacy_receipt_reference) <= 512)
);

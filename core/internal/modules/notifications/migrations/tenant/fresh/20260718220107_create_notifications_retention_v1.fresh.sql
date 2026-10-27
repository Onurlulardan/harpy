-- +goose Up
CREATE TABLE notifications_retention_checkpoints (
    singleton boolean PRIMARY KEY DEFAULT true CHECK (singleton),
    revision bigint NOT NULL CHECK (revision >= 0),
    last_job_id uuid,
    last_attempt_id uuid,
    last_completed_at timestamptz,
    CHECK ((revision = 0) = (last_job_id IS NULL)),
    CHECK ((last_job_id IS NULL) = (last_attempt_id IS NULL)),
    CHECK ((last_job_id IS NULL) = (last_completed_at IS NULL))
);

INSERT INTO notifications_retention_checkpoints (
    singleton, revision, last_job_id, last_attempt_id, last_completed_at
) VALUES (true, 0, NULL, NULL, NULL);

CREATE TABLE notifications_retention_receipts (
    job_id uuid PRIMARY KEY,
    attempt_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    payload_reference text COLLATE "C" NOT NULL
        CHECK (pg_catalog.octet_length(payload_reference) BETWEEN 1 AND 2048),
    payload_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(payload_sha256) = 32),
    successor_job_id uuid NOT NULL UNIQUE,
    successor_run_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(successor_run_at)),
    successor_dedupe_key text COLLATE "C" NOT NULL
        CHECK (pg_catalog.octet_length(successor_dedupe_key) BETWEEN 1 AND 512),
    content_redacted bigint NOT NULL CHECK (content_redacted >= 0),
    attempts_purged bigint NOT NULL CHECK (attempts_purged >= 0),
    receipts_purged bigint NOT NULL CHECK (receipts_purged >= 0),
    effects_purged bigint NOT NULL CHECK (effects_purged >= 0),
    deliveries_purged bigint NOT NULL CHECK (deliveries_purged >= 0),
    messages_purged bigint NOT NULL CHECK (messages_purged >= 0),
    hints_purged bigint NOT NULL CHECK (hints_purged >= 0),
    platform_records_purged bigint NOT NULL CHECK (platform_records_purged >= 0),
    has_remaining boolean NOT NULL,
    checkpoint_revision bigint NOT NULL CHECK (checkpoint_revision > 0),
    receipt_sha256 bytea NOT NULL CHECK (pg_catalog.octet_length(receipt_sha256) = 32),
    completed_at timestamptz NOT NULL CHECK (pg_catalog.isfinite(completed_at)),
    CHECK (successor_run_at > completed_at)
);

CREATE INDEX notifications_messages_dismissed_retention_idx
    ON notifications_messages (dismissed_at, id)
    WHERE purged_at IS NULL AND dismissed_at IS NOT NULL;
CREATE INDEX notifications_messages_read_retention_idx
    ON notifications_messages (read_at, id)
    WHERE purged_at IS NULL AND dismissed_at IS NULL AND read_at IS NOT NULL;
CREATE INDEX notifications_messages_unread_retention_idx
    ON notifications_messages (created_at, id)
    WHERE purged_at IS NULL AND dismissed_at IS NULL AND read_at IS NULL;
CREATE INDEX notifications_messages_tombstone_retention_idx
    ON notifications_messages (purged_at, id) WHERE purged_at IS NOT NULL;
CREATE INDEX notifications_effect_receipts_retention_idx
    ON notifications_effect_receipts (retain_until, message_id);
CREATE INDEX notifications_delivery_attempts_retention_idx
    ON notifications_delivery_attempts (finished_at, id);
CREATE INDEX notifications_delivery_receipts_retention_idx
    ON notifications_delivery_receipts (received_at, id);
CREATE INDEX notifications_realtime_hints_retention_idx
    ON notifications_realtime_hints (created_at, id);

GRANT SELECT ON TABLE notifications_retention_checkpoints TO forgecrud_runtime;
GRANT UPDATE (revision, last_job_id, last_attempt_id, last_completed_at) ON notifications_retention_checkpoints TO forgecrud_runtime;
GRANT SELECT ON TABLE notifications_retention_receipts TO forgecrud_runtime;
GRANT INSERT (job_id, attempt_id, tenant_id, payload_reference, payload_sha256, successor_job_id, successor_run_at, successor_dedupe_key, content_redacted, attempts_purged, receipts_purged, effects_purged, deliveries_purged, messages_purged, hints_purged, platform_records_purged, has_remaining, checkpoint_revision, receipt_sha256, completed_at) ON notifications_retention_receipts TO forgecrud_runtime;

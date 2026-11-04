-- +goose Up
CREATE TABLE notifications_retention_receipts (
    job_id uuid NOT NULL PRIMARY KEY,
    attempt_id uuid NOT NULL,
    tenant_id uuid NOT NULL,
    payload_reference text NOT NULL,
    payload_sha256 text NOT NULL,
    successor_job_id uuid NOT NULL,
    successor_run_at timestamptz,
    successor_dedupe_key text NOT NULL,
    content_redacted text NOT NULL,
    attempts_purged text NOT NULL,
    receipts_purged text NOT NULL,
    effects_purged text NOT NULL,
    deliveries_purged text NOT NULL,
    messages_purged text NOT NULL,
    hints_purged text NOT NULL,
    platform_records_purged text NOT NULL,
    has_remaining boolean NOT NULL,
    checkpoint_revision bigint NOT NULL,
    receipt_sha256 text NOT NULL,
    completed_at timestamptz
);

-- +goose Down
DROP TABLE notifications_retention_receipts;

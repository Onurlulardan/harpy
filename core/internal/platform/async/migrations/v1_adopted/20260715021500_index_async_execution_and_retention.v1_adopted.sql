-- +goose Up
CREATE INDEX jobs_claim_idx ON platform.jobs (
    scope_kind, scope_id, authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, priority DESC, run_at, id
) WHERE state = 'queued' AND parked_at IS NULL;

CREATE INDEX jobs_expired_lease_idx ON platform.jobs (
    scope_kind, scope_id, authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, lease_until, id
) WHERE state = 'running';

CREATE INDEX jobs_retention_idx ON platform.jobs (retain_until, id)
WHERE state IN ('succeeded', 'dead', 'cancelled');

CREATE INDEX jobs_payload_redaction_idx ON platform.jobs (completed_at, id)
WHERE state IN ('succeeded', 'dead', 'cancelled') AND payload_ref IS NOT NULL;

CREATE INDEX job_attempts_open_idx ON platform.job_attempts (job_id, attempt)
WHERE finished_at IS NULL;

CREATE INDEX job_attempts_effect_receipt_idx
ON platform.job_attempts (effect_receipt_id)
WHERE effect_receipt_id IS NOT NULL;

CREATE INDEX outbox_claim_idx ON platform.outbox (
    scope_kind, scope_id, authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, next_attempt_at, event_sequence, event_id
) WHERE delivery_state = 'pending';

CREATE INDEX outbox_expired_lease_idx ON platform.outbox (lease_until, event_id)
WHERE delivery_state = 'publishing';

CREATE INDEX outbox_stream_barrier_idx ON platform.outbox (
    scope_kind, scope_id, aggregate_type, aggregate_id, event_sequence
) WHERE delivery_state <> 'delivered';

CREATE INDEX outbox_lag_idx ON platform.outbox (created_at, event_id)
WHERE delivery_state <> 'delivered';

CREATE INDEX outbox_payload_redaction_idx ON platform.outbox (delivered_at, event_id)
WHERE delivery_state = 'delivered' AND payload_canonical IS NOT NULL;

CREATE INDEX inbox_latest_watermark_idx ON platform.inbox (
    consumer_name, consumer_version, scope_kind, scope_id,
    aggregate_type, aggregate_id, through_sequence DESC
) WHERE record_kind = 'watermark';

CREATE INDEX inbox_retention_idx ON platform.inbox (recorded_at, id);

CREATE UNIQUE INDEX idempotency_http_key_unique
ON platform.idempotency_receipts (
    tenant_id, principal_id, operation_id, idempotency_key
) NULLS NOT DISTINCT
WHERE record_kind = 'http';

CREATE UNIQUE INDEX idempotency_effect_identity_unique
ON platform.idempotency_receipts (
    scope_kind, scope_id, effect_id, effect_revision
) WHERE record_kind = 'effect';

CREATE INDEX idempotency_effect_key_lookup_idx
ON platform.idempotency_receipts (
    scope_kind, scope_id, provider, operation, effect_key, completed_at
) WHERE record_kind = 'effect';

CREATE INDEX idempotency_retention_idx
ON platform.idempotency_receipts (retain_until, receipt_id);

CREATE INDEX dead_letters_payload_retention_idx
ON platform.dead_letters (full_payload_expires_at, id)
WHERE record_kind = 'root' AND payload_canonical IS NOT NULL;

CREATE INDEX dead_letters_metadata_retention_idx
ON platform.dead_letters (metadata_expires_at, root_id, entry_sequence DESC);

CREATE INDEX dead_letters_operator_idx
ON platform.dead_letters (tenant_id, recorded_at DESC, id);

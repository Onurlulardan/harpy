//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
)

type asyncOutboxRow struct {
	id          string
	tenantID    any
	scopeKind   string
	scopeID     string
	groupID     string
	domain      string
	aggregateID string
	sequence    int
	payload     []byte
	payloadHash string
	delivered   bool
	redacted    bool
}

func newAsyncOutboxRow(id string) asyncOutboxRow {
	payload := []byte(`{"value":1}`)
	return asyncOutboxRow{
		id: id, tenantID: asyncTenantA, scopeKind: "tenant", scopeID: asyncTenantA,
		groupID: "operations", domain: "workflow",
		aggregateID: "77777777-7777-4777-8777-777777777777",
		sequence:    1, payload: payload, payloadHash: asyncSHA256(payload),
	}
}

func insertAsyncOutbox(
	ctx context.Context,
	transaction *sql.Tx,
	row asyncOutboxRow,
) error {
	state := "pending"
	var confirmedAt, deliveredAt, redactedAt any
	if row.delivered {
		state = "delivered"
		confirmedAt = "2026-07-15T10:00:01Z"
		deliveredAt = "2026-07-15T10:00:02Z"
	}
	if row.redacted {
		row.payload = nil
		redactedAt = "2026-07-15T10:00:03Z"
	}
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.outbox (
    event_id, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    envelope_contract_version, event_type, event_version,
    aggregate_type, aggregate_id, aggregate_version, event_sequence,
    occurred_at, actor_id, correlation_id, causation_id,
    idempotency_key, business_effect_idempotency_key,
    payload_canonical, payload_sha256, payload_schema_sha256,
    payload_redacted_at, delivery_state, delivery_attempts,
    lease_generation, next_attempt_at, confirmed_at, delivered_at,
    confirm_sha256, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, 1, $6, 1, 'v2',
    'integration-event.v1', 'workflow.changed', 1,
    'workflow.instance', $7, 1, $8,
    TIMESTAMPTZ '2026-07-15 10:00:00+00', $9, 'corr.outbox.1', $10,
    'event-key-1', 'business-effect-1', $11, $12, repeat('b', 64),
    $13, $14, 0, 0, TIMESTAMPTZ '2026-07-15 10:00:00+00',
    $15, $16,
    CASE WHEN $14 = 'delivered' THEN repeat('c', 64) ELSE NULL END,
    TIMESTAMPTZ '2026-07-15 10:00:00+00',
    TIMESTAMPTZ '2026-07-15 10:00:00+00'
)`, row.id, row.tenantID, row.scopeKind, row.scopeID, row.groupID, row.domain,
		row.aggregateID, row.sequence, asyncActor, asyncCausation, row.payload,
		row.payloadHash, redactedAt, state, confirmedAt, deliveredAt)
	return err
}

func asyncSHA256(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

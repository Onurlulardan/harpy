//go:build integration

package postgres

import (
	"context"
	"database/sql"
)

func insertAsyncDeadLetterRoot(
	ctx context.Context,
	transaction *sql.Tx,
	id string,
	tenantID any,
	scopeKind string,
	scopeID string,
) error {
	payload := []byte(`{"event":"failed"}`)
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.dead_letters (
    id, root_id, entry_sequence, record_kind,
    tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
	    source_kind, source_id, source_contract, source_version, source_sha256,
	    consumer_name, consumer_version,
	    reason_class, reason_code, redacted_error_sha256,
    payload_canonical, payload_sha256, redacted_envelope,
    full_payload_expires_at, metadata_expires_at, recorded_at
) VALUES (
    $1, $1, 0, 'root', $2, $3, $4,
    'operations', 1, 'workflow', 1, 'v2',
	    'event', $5, 'workflow.changed', 1, repeat('5', 64),
	    'projection.workflow', 1,
	    'permanent', 'EVENT_FAILED', repeat('6', 64),
    $6, $7, '{"event_type":"workflow.changed"}',
    TIMESTAMPTZ '2026-08-15 10:00:00+00',
    TIMESTAMPTZ '2027-07-15 10:00:00+00',
    TIMESTAMPTZ '2026-07-15 10:00:00+00'
)`, id, tenantID, scopeKind, scopeID, asyncCausation, payload,
		asyncSHA256(payload))
	return err
}

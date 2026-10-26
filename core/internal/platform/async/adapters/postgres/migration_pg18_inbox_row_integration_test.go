//go:build integration

package postgres

import (
	"context"
	"database/sql"
)

type asyncInboxScope struct {
	tenantID  any
	scopeKind string
	scopeID   string
}

func asyncTenantInboxScope() asyncInboxScope {
	return asyncInboxScope{
		tenantID: asyncTenantA, scopeKind: "tenant", scopeID: asyncTenantA,
	}
}

func insertAsyncInboxEvent(
	ctx context.Context,
	transaction *sql.Tx,
	id string,
	eventID string,
	consumerVersion int,
	sequence int,
	scope asyncInboxScope,
) error {
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.inbox (
    id, record_kind, consumer_name, consumer_version,
    tenant_id, scope_kind, scope_id,
    producer_group_id, producer_group_generation,
    producer_domain, producer_generation, writer_origin,
    consumer_group_id, consumer_group_generation,
    consumer_domain, consumer_generation, consumer_writer_origin,
    aggregate_type, aggregate_id, event_id, event_type,
    event_version, event_sequence, receipt_sha256, recorded_at
) VALUES (
    $1, 'event_receipt', 'projection.workflow', $2,
    $3, $4, $5, 'operations', 1, 'workflow', 1, 'v2',
    'operations', 1, 'workflow', 1, 'v2',
    'workflow.instance', '77777777-7777-4777-8777-777777777777',
    $6, 'workflow.changed', 1, $7, repeat('d', 64),
    TIMESTAMPTZ '2026-07-15 10:00:00+00'
)`, id, consumerVersion, scope.tenantID, scope.scopeKind, scope.scopeID,
		eventID, sequence)
	return err
}

func insertAsyncInboxBootstrap(
	ctx context.Context,
	transaction *sql.Tx,
	id string,
	causationID any,
) error {
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.inbox (
    id, record_kind, consumer_name, consumer_version,
    tenant_id, scope_kind, scope_id,
    producer_group_id, producer_group_generation,
    producer_domain, producer_generation, writer_origin,
    consumer_group_id, consumer_group_generation,
    consumer_domain, consumer_generation, consumer_writer_origin,
    aggregate_type, aggregate_id, through_sequence,
    watermark_kind, causation_id, state_sha256, receipt_sha256, recorded_at
) VALUES (
    $1, 'watermark', 'projection.workflow', 1,
    $2, 'tenant', $2, 'operations', 1, 'workflow', 1, 'v2',
    'operations', 1, 'workflow', 1, 'v2',
    'workflow.instance', '77777777-7777-4777-8777-777777777777', 5,
    'bootstrap', $3, repeat('e', 64), repeat('f', 64),
    TIMESTAMPTZ '2026-07-15 10:00:00+00'
)`, id, asyncTenantA, causationID)
	return err
}

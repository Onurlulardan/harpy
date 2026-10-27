//go:build integration

package postgres

import (
	"context"
	"database/sql"
)

func insertAsyncHTTPReceipt(
	ctx context.Context,
	transaction *sql.Tx,
	id string,
	tenantID any,
	scopeKind string,
	scopeID string,
	retentionClass string,
	asyncJobID any,
) error {
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.idempotency_receipts (
    receipt_id, record_kind, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    request_sha256, retention_class, principal_id, operation_id,
    idempotency_key, authorization_version, response_status,
    response_headers, response_body, response_sha256, async_job_id,
    completed_at, retain_until
) VALUES (
    $1, 'http', $2, $3, $4, 'operations', 1, 'workflow', 1, 'v2',
    repeat('1', 64), $5, $6, 'workflow.execute',
    'http-key-1', 1, 202, '{"content-type":"application/json"}',
    convert_to('{"accepted":true}', 'UTF8'), repeat('2', 64), $7,
    TIMESTAMPTZ '2026-07-15 10:00:00+00',
    TIMESTAMPTZ '2026-10-15 10:00:00+00'
)`, id, tenantID, scopeKind, scopeID, retentionClass, asyncActor, asyncJobID)
	return err
}

func insertAsyncEffectReceipt(
	ctx context.Context,
	transaction *sql.Tx,
	id string,
	tenantID any,
	scopeKind string,
	scopeID string,
	effectKind any,
	provider any,
) error {
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.idempotency_receipts (
    receipt_id, record_kind, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
	    request_sha256, retention_class, effect_id, effect_kind, effect_revision,
	    effect_key, provider, operation, correlation_id, causation_id,
    effect_status, effect_code, effect_result_ref, effect_result_sha256,
    completed_at, retain_until
) VALUES (
	    $1, 'effect', $2, $3, $4, 'operations', 1, 'workflow', 1, 'v2',
	    repeat('3', 64), 'external_effect', $5, $6, 1,
	    'effect-key-1', $7, 'send', 'corr.effect.1', $8,
    'succeeded', 'DELIVERED', 'provider/result/1', repeat('4', 64),
    TIMESTAMPTZ '2026-07-15 10:00:00+00',
    TIMESTAMPTZ '2026-10-15 10:00:00+00'
)`, id, tenantID, scopeKind, scopeID, asyncCausation, effectKind, provider, asyncActor)
	return err
}

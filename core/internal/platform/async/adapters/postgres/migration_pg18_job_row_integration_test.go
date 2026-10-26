//go:build integration

package postgres

import (
	"context"
	"database/sql"
)

func insertAsyncQueuedJob(
	ctx context.Context,
	transaction *sql.Tx,
	id string,
	tenantID any,
	scopeKind string,
	scopeID string,
	groupID string,
	domain string,
	dedupeKey string,
) error {
	_, err := transaction.ExecContext(ctx, `
INSERT INTO platform.jobs (
    id, tenant_id, scope_kind, scope_id,
    authority_group_id, authority_group_generation,
    authority_domain, cutover_generation, writer_origin,
    kind, payload_revision, payload_ref, payload_sha256, dedupe_key,
    effect_class, priority, run_at, state, attempts, max_attempts,
    lease_generation, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, 1, $6, 1, 'v2',
    'workflow.execute', 1, 'payload/job/1', repeat('a', 64), $7,
    'database_transactional', 0,
    TIMESTAMPTZ '2026-07-15 10:00:00+00', 'queued', 0, 8, 0,
    TIMESTAMPTZ '2026-07-15 09:00:00+00',
    TIMESTAMPTZ '2026-07-15 09:00:00+00'
)`, id, tenantID, scopeKind, scopeID, groupID, domain, dedupeKey)
	return err
}

//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const (
	iamLifecycleReplaySession   = "7a000000-0000-4000-8000-000000000001"
	iamLifecycleRaceSession     = "7a000000-0000-4000-8000-000000000002"
	iamLifecycleStaleSession    = "7a000000-0000-4000-8000-000000000003"
	iamLifecycleRollbackSession = "7a000000-0000-4000-8000-000000000004"
	iamLifecycleReplayOperation = "7b000000-0000-4000-8000-000000000001"
	iamLifecycleRaceOperation   = "7b000000-0000-4000-8000-000000000002"
	iamLifecycleStaleOperation  = "7b000000-0000-4000-8000-000000000003"
	iamLifecycleRollbackOp      = "7b000000-0000-4000-8000-000000000004"
)

const insertIAMLifecycleSession = `
INSERT INTO iam_sessions (
    id, user_id, access_mode_id, access_mode, portal_binding_id,
    state, device_kind, browser_family, os_family, ip_hmac_key_id,
    ip_hmac, revision, created_at, updated_at, last_seen_at,
    absolute_expires_at, revoked_at, expired_at, retain_until
) VALUES (
    $1::uuid, '51000000-0000-4000-8000-000000000001',
    '52000000-0000-4000-8000-000000000001', 'operator_console', NULL,
    'active', 'desktop', 'chromium', 'windows', 'fixture.v1',
    pg_catalog.decode(pg_catalog.repeat('71', 32), 'hex'), $2::bigint,
	pg_catalog.statement_timestamp() - INTERVAL '1 hour',
	pg_catalog.statement_timestamp() - INTERVAL '1 hour',
	pg_catalog.statement_timestamp() - INTERVAL '1 hour',
	pg_catalog.statement_timestamp() + INTERVAL '14 days', NULL, NULL,
	pg_catalog.statement_timestamp() + INTERVAL '60 days'
)`

func seedIAMLifecycleRuntimeSessions(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	fixtures := []struct {
		id       string
		revision int64
	}{
		{iamLifecycleReplaySession, 1},
		{iamLifecycleRaceSession, 1},
		{iamLifecycleStaleSession, 2},
		{iamLifecycleRollbackSession, 1},
	}
	for _, tenant := range iamFoundationTenants {
		for _, fixture := range fixtures {
			mustExecIAMTenant(t, ctx, database, tenant.schema,
				insertIAMLifecycleSession, fixture.id, fixture.revision)
		}
	}
}

//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const iamRBACRuntimeAuthorityBindingID = "97000000-0000-4000-8000-000000000001"

func seedIAMRBACRuntimeAuthority(t *testing.T, ctx context.Context, database *sql.DB, schema string) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, `
UPDATE iam_users SET authorization_version=1, revision=revision+1,
    updated_at='2026-07-16T00:20:00Z'
WHERE id='51000000-0000-4000-8000-000000000001';
INSERT INTO iam_user_permission_bindings (
    id, user_id, permission_id, permission_match_kind, state, revision,
    created_at, updated_at, activated_at, revoked_at
) VALUES (
    '97000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', 'exact', 'active', 1,
    '2026-07-16T00:20:00Z', '2026-07-16T00:20:00Z',
    '2026-07-16T00:20:00Z', NULL
);`)
}

func iamRBACActorVersions(t *testing.T, ctx context.Context, database *sql.DB,
	schema string) (int64, int64) {
	t.Helper()
	var tokenVersion, authorizationVersion int64
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT token_version, authorization_version FROM iam_users
WHERE id='51000000-0000-4000-8000-000000000001'`,
		&tokenVersion, &authorizationVersion)
	return tokenVersion, authorizationVersion
}

func iamRBACClock(t *testing.T, ctx context.Context, database *sql.DB, schema string) int64 {
	t.Helper()
	var clock int64
	mustQueryIAMTenantValue(t, ctx, database, schema,
		"SELECT clock_value FROM iam_authorization_clock WHERE singleton", &clock)
	return clock
}

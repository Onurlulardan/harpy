//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

const updateIAMRBACProtectedPolicy = `
SELECT pg_catalog.concat_ws('|', result_id::text, result_revision::text,
    authorization_clock::text, replayed::text)
FROM iam_rbac_protected_admin_update_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea,
    '94000000-0000-4000-8000-000000000001', $3::integer, 1,
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', 1, 2
)`

func assertIAMRBACProtectedPolicyRuntime(t *testing.T, ctx context.Context,
	database *sql.DB, schema string) {
	t.Helper()
	_, err := runIAMAtomicText(ctx, database, schema, updateIAMRBACProtectedPolicy,
		"99300000-0000-4000-8000-000000000001", bytes.Repeat([]byte{0x71}, 32), 2)
	if err == nil {
		t.Fatal("protected admin minimum accepted unavailable assignments")
	}
	if iamRBACClock(t, ctx, database, schema) != 4 {
		t.Fatal("failed protected admin mutation advanced the clock")
	}
	result := mustRunIAMAtomicText(t, ctx, database, schema, updateIAMRBACProtectedPolicy,
		"99300000-0000-4000-8000-000000000002", bytes.Repeat([]byte{0x72}, 32), 1)
	if result != iamRBACProtectedRoleID+"|2|5|false" {
		t.Fatalf("protected admin update = %s", result)
	}
	policy := mustRunIAMAtomicText(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|', role_id::text,
    minimum_active_assignments::text, authorization_epoch::text, revision::text)
FROM iam_rbac_protected_admin_get_v1(
    pg_catalog.current_schema()::name,
    '94000000-0000-4000-8000-000000000001',
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001')`)
	if policy != iamRBACProtectedRoleID+"|1|5|2" {
		t.Fatalf("protected admin read = %s", policy)
	}
}

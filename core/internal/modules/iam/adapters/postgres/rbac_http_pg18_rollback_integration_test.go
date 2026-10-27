//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
)

func assertIAMRBACHTTPRollback(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	mustExecIAMTenant(t, ctx, database, schema, `
INSERT INTO iam_actions (
 id, action_key, display_name, description, is_protected,
 state, revision, created_at, updated_at, retired_at
) VALUES (
 'c1000000-0000-4000-8000-000000000001', 'rollback_read',
 'Rollback Read', NULL, false, 'active', 1,
 '2026-07-16T00:30:00Z', '2026-07-16T00:30:00Z', NULL
)`)
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	items := `[{
 "ordinal":1,
 "action_id":"c1000000-0000-4000-8000-000000000001",
 "permission_id":"c2000000-0000-4000-8000-000000000001",
 "binding_id":"c3000000-0000-4000-8000-000000000001"
}]`
	purpose := sha256.Sum256([]byte(
		"harpy/iam/rbac-step-up/v1|iam.rbac.assignment.create|",
	))
	var result string
	err = transaction.QueryRowContext(
		ctx, createIAMRBACHTTPAssignment,
		"c5000000-0000-4000-8000-000000000001",
		bytes.Repeat([]byte{0xf6}, 32),
		"c4000000-0000-4000-8000-000000000001",
		"USER", iamUserThreeID, items, int64(1), int64(1), purpose[:],
	).Scan(&result)
	if err != nil || result != "c4000000-0000-4000-8000-000000000001|1|4|false|true" {
		_ = transaction.Rollback()
		t.Fatalf("rollback assignment result=%s error=%v", result, err)
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	var artifacts int
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT
 (SELECT pg_catalog.count(*) FROM iam_rbac_assignment_groups
  WHERE id='c4000000-0000-4000-8000-000000000001') +
 (SELECT pg_catalog.count(*) FROM iam_rbac_assignment_members
  WHERE group_id='c4000000-0000-4000-8000-000000000001') +
 (SELECT pg_catalog.count(*) FROM iam_permissions
  WHERE id='c2000000-0000-4000-8000-000000000001') +
 (SELECT pg_catalog.count(*) FROM iam_user_permission_bindings
  WHERE id='c3000000-0000-4000-8000-000000000001') +
 (SELECT pg_catalog.count(*) FROM iam_authorization_operation_receipts
  WHERE operation_id='c5000000-0000-4000-8000-000000000001') +
 (SELECT pg_catalog.count(*) FROM iam_rbac_change_audit
  WHERE operation_id='c5000000-0000-4000-8000-000000000001')`, &artifacts)
	if artifacts != 0 || iamRBACClock(t, ctx, database, schema) != 3 {
		t.Fatalf("rollback artifacts=%d clock=%d", artifacts,
			iamRBACClock(t, ctx, database, schema))
	}
}

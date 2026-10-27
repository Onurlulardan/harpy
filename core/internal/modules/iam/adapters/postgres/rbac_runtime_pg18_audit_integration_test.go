//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func assertIAMRBACRuntimeAuditRollback(t *testing.T, ctx context.Context,
	database *sql.DB, schema string) {
	t.Helper()
	operationID := "98300000-0000-4000-8000-000000000001"
	actionID := "98400000-0000-4000-8000-000000000001"
	hash := bytes.Repeat([]byte{0x43}, 32)
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var revision, clock int64
	var replayed bool
	err = transaction.QueryRowContext(ctx, `
SELECT result_revision, authorization_clock, replayed
FROM iam_rbac_action_create_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea, $3::uuid,
    'rollback_action', 'Rollback Action', 'rollback test', false,
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', 1, 2)`,
		operationID, hash, actionID).Scan(&revision, &clock, &replayed)
	if err != nil || revision != 1 || clock != 2 || replayed {
		_ = transaction.Rollback()
		t.Fatalf("rollback mutation result = %d/%d/%t, error = %v", revision, clock, replayed, err)
	}
	var written bool
	err = transaction.QueryRowContext(ctx, `SELECT iam_rbac_audit_write_v1(
    pg_catalog.current_schema()::name, $1::uuid,
    '51000000-0000-4000-8000-000000000001', $2::uuid, $3::bytea,
    'iam.rbac.action.create', 'iam.rbac.action.created', 1, 2)`,
		operationID, actionID, hash).Scan(&written)
	if err != nil || !written {
		_ = transaction.Rollback()
		t.Fatalf("audit write failed: %v", err)
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	var artifacts int
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT (SELECT pg_catalog.count(*) FROM iam_actions WHERE id='98400000-0000-4000-8000-000000000001') +
       (SELECT pg_catalog.count(*) FROM iam_authorization_operation_receipts WHERE operation_id='98300000-0000-4000-8000-000000000001') +
       (SELECT pg_catalog.count(*) FROM iam_rbac_change_audit WHERE operation_id='98300000-0000-4000-8000-000000000001')`, &artifacts)
	if artifacts != 0 || iamRBACClock(t, ctx, database, schema) != 1 {
		t.Fatal("sink rollback left RBAC state, receipt, audit, or clock evidence")
	}
}

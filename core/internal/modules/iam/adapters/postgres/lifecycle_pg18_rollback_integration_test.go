//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func assertIAMLifecycleRollback(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, tenant.schema, iamUserOneID)
	arguments := lifecycleRevokeArguments(tenant, iamLifecycleRollbackOp,
		bytes.Repeat([]byte{0x76}, 32), iamLifecycleRollbackSession, 1, actor)
	connection, transaction, err := beginIAMAtomicRuntimeTx(
		ctx, database, tenant.schema,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var result string
	err = transaction.QueryRowContext(ctx, revokeIAMLifecycleSession,
		arguments...).Scan(&result)
	if err != nil || result != iamLifecycleRollbackSession+"|2|false" {
		_ = transaction.Rollback()
		t.Fatalf("%s lifecycle rollback probe=%q: %v", tenant.schema, result, err)
	}
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	assertIAMLifecycleState(t, ctx, database, tenant.schema,
		iamLifecycleRollbackSession, "active", 1)
	assertIAMLifecycleReceiptCount(t, ctx, database, tenant.schema,
		iamLifecycleRollbackOp, 0)
	var audit int
	mustQueryIAMTenantValue(t, ctx, database, tenant.schema, `
SELECT pg_catalog.count(*) FROM iam_lifecycle_audit
WHERE operation_id='7b000000-0000-4000-8000-000000000004'`, &audit)
	if audit != 0 {
		t.Fatalf("%s lifecycle rollback audit=%d", tenant.schema, audit)
	}
}

func assertIAMLifecycleState(
	t *testing.T, ctx context.Context, database *sql.DB,
	schema, sessionID, wantState string, wantRevision int64,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	var state string
	var revision int64
	err := transaction.QueryRowContext(ctx, `
SELECT state,revision FROM iam_sessions WHERE id=$1`, sessionID).Scan(
		&state, &revision,
	)
	_ = transaction.Rollback()
	if err != nil || state != wantState || revision != wantRevision {
		t.Fatalf("%s session %s=%s/%d: %v",
			schema, sessionID, state, revision, err)
	}
}

func assertIAMLifecycleReceiptCount(
	t *testing.T, ctx context.Context, database *sql.DB,
	schema, operationID string, want int,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	var count int
	err := transaction.QueryRowContext(ctx, `
SELECT pg_catalog.count(*) FROM iam_lifecycle_operation_receipts
WHERE operation_id=$1`, operationID).Scan(&count)
	_ = transaction.Rollback()
	if err != nil || count != want {
		t.Fatalf("%s lifecycle receipt %s=%d want=%d: %v",
			schema, operationID, count, want, err)
	}
}

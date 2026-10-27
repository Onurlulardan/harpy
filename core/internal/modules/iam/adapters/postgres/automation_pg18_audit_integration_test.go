//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"testing"
)

func assertIAMServicePrincipalAudit(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	index int,
) {
	t.Helper()
	marker := fmt.Sprintf("%02x", 0x71+index)
	principalID := marker + "100000-0000-4000-8000-000000000001"
	operationID := marker + "200000-0000-4000-8000-000000000001"
	hash := bytes.Repeat([]byte{byte(0x61 + index)}, 32)
	tokenVersion, authorizationVersion := iamRBACActorVersions(
		t, ctx, database, schema,
	)
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var revision, clock int64
	var replayed bool
	err = transaction.QueryRowContext(ctx, `
SELECT result_revision, authorization_clock, replayed
FROM iam_service_principal_create_v1(
    pg_catalog.current_schema()::name, $1::uuid, $2::bytea, $3::uuid,
    'workflow.audit', 'Workflow audit',
    '51000000-0000-4000-8000-000000000001',
    '93000000-0000-4000-8000-000000000001', $4::bigint, $5::bigint,
    pg_catalog.decode(pg_catalog.repeat('a1',32),'hex'),
    pg_catalog.transaction_timestamp(), false, true, false)`,
		operationID, hash, principalID, tokenVersion, authorizationVersion,
	).Scan(&revision, &clock, &replayed)
	if err != nil || replayed || revision != 1 {
		_ = transaction.Rollback()
		t.Fatalf("service principal create = %d/%d/%t, %v", revision, clock, replayed, err)
	}
	var written bool
	err = transaction.QueryRowContext(ctx, `SELECT iam_rbac_audit_write_v1(
    pg_catalog.current_schema()::name, $1::uuid,
    '51000000-0000-4000-8000-000000000001', $2::uuid, $3::bytea,
    'iam.service_principal.create', 'iam.service_principal.created',
    $4::bigint, $5::bigint)`,
		operationID, principalID, hash, revision, clock,
	).Scan(&written)
	if err != nil || !written {
		_ = transaction.Rollback()
		t.Fatalf("service principal audit write = %t, %v", written, err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	queryConnection, queryTx := beginIAMTenantTx(t, ctx, database, schema)
	defer queryConnection.Close()
	var count int
	err = queryTx.QueryRowContext(ctx, `
SELECT pg_catalog.count(*) FROM iam_rbac_change_audit
WHERE operation_id=$1::uuid AND command_key='iam.service_principal.create'`,
		operationID).Scan(&count)
	if err != nil || count != 1 {
		_ = queryTx.Rollback()
		t.Fatalf("service principal audit count = %d, %v", count, err)
	}
	if err := queryTx.Commit(); err != nil {
		t.Fatal(err)
	}
}

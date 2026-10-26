//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func assertIAMIdentityChangeRollback(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	const operationID = "aa100000-0000-4000-8000-000000000001"
	const aggregateID = "aa200000-0000-4000-8000-000000000001"
	request := bytes.Repeat([]byte{0x61}, 32)
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)
	probeIAMIdentityRuntimeTx(t, ctx, database, "t_iam_a", func(tx *sql.Tx) {
		createIAMIdentityAuditOrganization(
			t, ctx, tx, operationID, aggregateID, request, actor,
		)
		var revision, clock, sequence int64
		if err := tx.QueryRowContext(ctx, iamIdentityChangeAuditWrite,
			operationID, iamIdentityActorID, aggregateID, request,
		).Scan(&revision, &clock, &sequence); err != nil {
			t.Fatal(err)
		}
	})
	var artifacts int
	mustQueryIAMTenantValue(t, ctx, database, "t_iam_a", `
SELECT (SELECT pg_catalog.count(*) FROM iam_organizations
        WHERE id='aa200000-0000-4000-8000-000000000001') +
       (SELECT pg_catalog.count(*) FROM iam_authorization_operation_receipts
        WHERE operation_id='aa100000-0000-4000-8000-000000000001') +
       (SELECT pg_catalog.count(*) FROM iam_identity_change_audit
        WHERE operation_id='aa100000-0000-4000-8000-000000000001')`, &artifacts)
	if artifacts != 0 {
		t.Fatalf("identity sink rollback left %d artifacts", artifacts)
	}
}

func assertIAMIdentityChangeACL(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	var tableDML, execute bool
	err := database.QueryRowContext(ctx, `
SELECT pg_catalog.has_table_privilege(
           'harpy_runtime','t_iam_a.iam_identity_change_audit',
           'SELECT,INSERT,UPDATE,DELETE,TRUNCATE'),
       pg_catalog.has_function_privilege(
           'harpy_runtime',
           't_iam_a.iam_identity_change_audit_write_v1(name,uuid,uuid,text,uuid,bytea,text,text)',
           'EXECUTE')`).Scan(&tableDML, &execute)
	if err != nil || tableDML || !execute {
		t.Fatalf("identity change ACL table=%t execute=%t error=%v",
			tableDML, execute, err)
	}
	if _, err := database.ExecContext(ctx, `
UPDATE t_iam_a.iam_identity_change_audit
SET recorded_at = pg_catalog.clock_timestamp()`); err == nil {
		t.Fatal("identity change audit accepted mutation")
	}
}

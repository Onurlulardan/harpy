//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"
)

func runIAMPermissionTransition(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	statement string,
	arguments []any,
	operationID, groupID, actorID string,
	reviewTask any,
	command string,
	fromState any,
	toState string,
	revision int64,
) string {
	t.Helper()
	connection, transaction, err := beginIAMPGFirstRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var result string
	if err := transaction.QueryRowContext(ctx, statement, arguments...).Scan(&result); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("permission transition mutation: %s", iamPGErrorDetail(err))
	}
	var clock int64
	err = transaction.QueryRowContext(ctx, `
SELECT iam_permission_request_transition_audit_write_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,$6,$7,$8,$9
)`, operationID, []byte(strings.Repeat("c", 32)), groupID, actorID,
		reviewTask, command, fromState, toState, revision).Scan(&clock)
	if err != nil || clock < 0 {
		_ = transaction.Rollback()
		t.Fatalf("permission transition audit: %s", iamPGErrorDetail(err))
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return result
}

func assertIAMPermissionTransitionSinkRollback(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	marker string,
) {
	t.Helper()
	groupID := iamPermissionInvitationID("c1", marker)
	requestID := iamPermissionInvitationID("c2", marker)
	bindingID := iamPermissionInvitationID("c3", marker)
	document := fmt.Sprintf(
		`[{"id":"%s","ordinal":1,"target_kind":"USER","target_id":"%s",`+
			`"permission_id":"%s","planned_binding_id":"%s"}]`,
		requestID, iamPermissionRequester, iamPermissionRead, bindingID,
	)
	connection, transaction, err := beginIAMPGFirstRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var created string
	err = transaction.QueryRowContext(ctx, `
SELECT group_id::text FROM iam_permission_request_group_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,'SELF_SERVICE',$3,$4,1,$5,$6,$7::jsonb
)`, groupID, iamPermissionRequester, iamPermissionManage, iamPermissionReview,
		[]byte(strings.Repeat("p", 32)), time.Now().UTC().Add(time.Hour), document).Scan(&created)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	_, err = transaction.ExecContext(ctx, `
SELECT iam_permission_request_transition_audit_write_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,NULL,'iam.permission_request.create',NULL,'PENDING',1
)`, iamPermissionInvitationID("c4", marker), []byte(strings.Repeat("c", 32)),
		groupID, iamPermissionReviewer)
	if err == nil {
		_ = transaction.Rollback()
		t.Fatal("invalid transition audit unexpectedly succeeded")
	}
	_ = transaction.Rollback()
	var count int64
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT pg_catalog.count(*) FROM iam_permission_request_groups
WHERE id = '`+groupID+`'`, &count)
	if count != 0 {
		t.Fatal("transition sink failure committed permission request state")
	}
}

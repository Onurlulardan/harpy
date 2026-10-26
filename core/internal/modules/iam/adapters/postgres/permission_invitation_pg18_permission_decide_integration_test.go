//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"
)

func decideIAMPermissionRequest(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	groupID string,
	marker string,
) {
	t.Helper()
	operationID := iamPermissionInvitationID("a5", marker)
	taskID := iamPermissionInvitationID("a4", marker)
	bindingID := iamPermissionInvitationID("a3", marker)
	arguments := []any{
		operationID, []byte(strings.Repeat("d", 32)), groupID,
		iamPermissionReviewer, taskID, []byte(strings.Repeat("p", 32)),
		time.Now().UTC().Add(-time.Minute),
	}
	rollbackIAMPGFirstText(t, ctx, database, schema, iamPermissionDecisionSQL, arguments...)
	assertPermissionDecisionState(
		t, ctx, database, schema, groupID, bindingID, "IN_REVIEW", 2, 0, 1,
	)
	substituted := append([]any(nil), arguments...)
	substituted[0] = iamPermissionInvitationID("a6", marker)
	substituted[5] = []byte(strings.Repeat("x", 32))
	assertIAMSQLState(
		t, rejectIAMPGFirst(t, ctx, database, schema, iamPermissionDecisionSQL, substituted...), "42501",
	)
	staleStepUp := append([]any(nil), arguments...)
	staleStepUp[0] = iamPermissionInvitationID("a7", marker)
	staleStepUp[6] = time.Now().UTC().Add(-11 * time.Minute)
	assertIAMSQLState(
		t, rejectIAMPGFirst(t, ctx, database, schema, iamPermissionDecisionSQL, staleStepUp...), "42501",
	)
	result := mustRunIAMPGFirstText(
		t, ctx, database, schema, iamPermissionDecisionSQL, arguments...,
	)
	if result != "iam.permission_request.approved|3|1|false" {
		t.Fatalf("permission request decision = %q", result)
	}
	assertPermissionDecisionState(
		t, ctx, database, schema, groupID, bindingID, "APPROVED", 3, 1, 2,
	)
	auditWritten := mustRunIAMPGFirstText(
		t, ctx, database, schema, iamPermissionDecisionAuditSQL,
		operationID, arguments[1], groupID, iamPermissionReviewer, taskID,
	)
	if auditWritten != "true" {
		t.Fatalf("permission decision audit written = %q", auditWritten)
	}
	replayed := mustRunIAMPGFirstText(
		t, ctx, database, schema, iamPermissionDecisionSQL, arguments...,
	)
	if replayed != "iam.permission_request.approved|3|1|true" {
		t.Fatalf("permission request replay = %q", replayed)
	}
	if countIAMPermissionDecisionAudit(t, ctx, database, schema, operationID) != 1 {
		t.Fatal("permission decision replay duplicated immutable audit")
	}
	staleOperation := iamPermissionInvitationID("a9", marker)
	staleArgs := append([]any(nil), arguments...)
	staleArgs[0] = staleOperation
	err := rejectIAMPGFirst(t, ctx, database, schema, iamPermissionDecisionSQL, staleArgs...)
	assertIAMSQLState(t, err, "40001")
}

const iamPermissionDecisionSQL = `
SELECT pg_catalog.concat_ws('|',result_code,result_revision::text,
    authorization_clock::text,replayed::text)
FROM iam_permission_request_decide_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,2,'IN_REVIEW',$4,1,1,$5,'APPROVE',
    $6,$7,'urn:harpy:acr:step-up',true,false,false
)`

const iamPermissionDecisionAuditSQL = `
SELECT iam_permission_request_audit_write_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,$4,$5,'APPROVE','iam.permission_request.approved',3,1
)::text`

func countIAMPermissionDecisionAudit(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	operationID string,
) int64 {
	t.Helper()
	var count int64
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT pg_catalog.count(*)
FROM iam_permission_request_decision_audit
WHERE operation_id = '`+operationID+`'`, &count)
	return count
}

func assertPermissionDecisionState(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	groupID string,
	bindingID string,
	wantState string,
	wantRevision int64,
	wantClock int64,
	wantUserVersion int64,
) {
	t.Helper()
	var state string
	var revision, clock, bindingCount, userVersion int64
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT request_group.state, request_group.revision, clock_state.clock_value,
       (SELECT pg_catalog.count(*) FROM iam_user_permission_bindings AS binding
        WHERE binding.id = '`+bindingID+`'), principal.authorization_version
FROM iam_permission_request_groups AS request_group
CROSS JOIN iam_authorization_clock AS clock_state
JOIN iam_users AS principal
  ON principal.id = '51000000-0000-4000-8000-000000000002'
WHERE request_group.id = '`+groupID+`' AND clock_state.singleton`,
		&state, &revision, &clock, &bindingCount, &userVersion)
	if state != wantState || revision != wantRevision || clock != wantClock ||
		userVersion != wantUserVersion {
		t.Fatalf(
			"permission decision state = %s/%d clock=%d user=%d",
			state, revision, clock, userVersion,
		)
	}
	wantBindingCount := int64(0)
	if wantState == "APPROVED" {
		wantBindingCount = 1
	}
	if bindingCount != wantBindingCount {
		t.Fatalf("permission decision binding count = %d", bindingCount)
	}
}

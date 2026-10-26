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

func assertIAMPermissionRequestGroupAtomicity(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	marker string,
) {
	t.Helper()
	groupID := iamPermissionInvitationID("e1", marker)
	roleBindingID := iamPermissionInvitationID("e4", marker)
	document := permissionAtomicityDocument(marker, roleBindingID)
	mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT group_state FROM iam_permission_request_group_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,'PRIVILEGED',$3,$4,1,$5,$6,$7::jsonb
)`, groupID, iamPermissionReviewer, iamPermissionManage, iamPermissionReview,
		[]byte(strings.Repeat("g", 32)), time.Now().UTC().Add(time.Hour), document)
	taskID := iamPermissionInvitationID("e6", marker)
	mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT group_state FROM iam_permission_request_review_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,1,'PENDING',$2,1,$3
)`, groupID, iamPermissionReviewer, taskID)
	err := rejectIAMPGFirst(t, ctx, database, schema, `
SELECT result_code FROM iam_permission_request_decide_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$3,2,'IN_REVIEW',$4,1,1,$5,'APPROVE',
    $6,$7,'urn:harpy:acr:step-up',true,false,false
)`, iamPermissionInvitationID("e7", marker),
		[]byte(strings.Repeat("r", 32)), groupID, iamPermissionReviewer, taskID,
		[]byte(strings.Repeat("g", 32)), time.Now().UTC().Add(-time.Minute))
	assertIAMSQLState(t, err, "40001")
	assertIAMPermissionRequestAtomicRollback(
		t, ctx, database, schema, groupID, roleBindingID,
	)
}

func permissionAtomicityDocument(marker string, roleBindingID string) string {
	return fmt.Sprintf(`[
{"id":"%s","ordinal":1,"target_kind":"ROLE","target_id":"%s",`+
		`"permission_id":"%s","planned_binding_id":"%s"},
{"id":"%s","ordinal":2,"target_kind":"USER","target_id":"%s",`+
		`"permission_id":"%s","planned_binding_id":"%s"}
]`, iamPermissionInvitationID("e2", marker), iamInvitationRole,
		iamPermissionRead, roleBindingID,
		iamPermissionInvitationID("e3", marker), iamInvitationAcceptUser,
		iamPermissionRead, iamPermissionInvitationID("e5", marker))
}

func assertIAMPermissionRequestAtomicRollback(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	groupID string,
	roleBindingID string,
) {
	t.Helper()
	var state string
	var revision, bindingCount, receiptCount, clock int64
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT request_group.state, request_group.revision,
       (SELECT pg_catalog.count(*) FROM iam_role_permission_bindings AS binding
        WHERE binding.id = '`+roleBindingID+`'),
       (SELECT pg_catalog.count(*) FROM iam_authorization_operation_receipts
        WHERE result_id = '`+groupID+`'), clock_state.clock_value
FROM iam_permission_request_groups AS request_group
CROSS JOIN iam_authorization_clock AS clock_state
WHERE request_group.id = '`+groupID+`' AND clock_state.singleton`,
		&state, &revision, &bindingCount, &receiptCount, &clock)
	if state != "IN_REVIEW" || revision != 2 || bindingCount != 0 ||
		receiptCount != 0 || clock != 1 {
		t.Fatalf(
			"permission group rollback = %s/%d binding=%d receipt=%d clock=%d",
			state, revision, bindingCount, receiptCount, clock,
		)
	}
}

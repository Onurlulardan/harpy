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

const (
	iamPermissionRequester = "51000000-0000-4000-8000-000000000002"
	iamPermissionReviewer  = "51000000-0000-4000-8000-000000000001"
	iamPermissionRead      = "93000000-0000-4000-8000-000000000001"
	iamPermissionManage    = "a0030000-0000-4000-8000-000000000001"
	iamPermissionReview    = "a0030000-0000-4000-8000-000000000002"
)

func assertIAMPermissionRequestRuntime(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	marker := "8a"
	if schema == "t_iam_b" {
		marker = "8b"
	}
	approve := createIAMPermissionRequest(t, ctx, database, schema, marker, "a1", "a2", "a3")
	cancel := createIAMPermissionRequest(t, ctx, database, schema, marker, "a6", "a7", "a8")
	assertIAMPermissionRequestReads(t, ctx, database, schema, approve)
	cancelIAMPermissionRequest(t, ctx, database, schema, cancel, marker)
	expireIAMPermissionRequest(t, ctx, database, schema, marker)
	reviewIAMPermissionRequest(t, ctx, database, schema, approve, marker)
	assertIAMPermissionTransitionSinkRollback(t, ctx, database, schema, marker)
	decideIAMPermissionRequest(t, ctx, database, schema, approve, marker)
	assertIAMPermissionRequestGroupAtomicity(t, ctx, database, schema, marker)
}

func createIAMPermissionRequest(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	marker string,
	groupPrefix string,
	requestPrefix string,
	bindingPrefix string,
) string {
	t.Helper()
	groupID := iamPermissionInvitationID(groupPrefix, marker)
	requestID := iamPermissionInvitationID(requestPrefix, marker)
	bindingID := iamPermissionInvitationID(bindingPrefix, marker)
	document := fmt.Sprintf(
		`[{"id":"%s","ordinal":1,"target_kind":"USER","target_id":"%s",`+
			`"permission_id":"%s","planned_binding_id":"%s"}]`,
		requestID, iamPermissionRequester, iamPermissionRead, bindingID,
	)
	result := runIAMPermissionTransition(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|',group_id::text,group_state,group_revision::text)
FROM iam_permission_request_group_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,'SELF_SERVICE',$3,$4,1,$5,$6,$7::jsonb
)`, []any{groupID, iamPermissionRequester, iamPermissionManage, iamPermissionReview,
		[]byte(strings.Repeat("p", 32)), time.Now().UTC().Add(time.Hour), document},
		groupID, groupID, iamPermissionRequester, nil,
		"iam.permission_request.create", nil, "PENDING", 1)
	if result != groupID+"|PENDING|1" {
		t.Fatalf("permission request create = %q", result)
	}
	return groupID
}

func assertIAMPermissionRequestReads(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	groupID string,
) {
	t.Helper()
	read := mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|',group_id::text,group_state,group_revision::text)
FROM iam_permission_request_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2
)`, iamPermissionRequester, groupID)
	if read != groupID+"|PENDING|1" {
		t.Fatalf("permission request read = %q", read)
	}
	listed := mustRunIAMPGFirstText(t, ctx, database, schema, `
SELECT pg_catalog.count(*)::text FROM iam_permission_request_list_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,'["PENDING"]'::jsonb,NULL,101
)`, iamPermissionRequester)
	if listed != "3" {
		t.Fatalf("pending permission request count = %s, want 3", listed)
	}
}

func cancelIAMPermissionRequest(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	groupID string,
	marker string,
) {
	t.Helper()
	result := runIAMPermissionTransition(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|',group_id::text,group_state,group_revision::text)
FROM iam_permission_request_cancel_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,1,'PENDING',$2,1
)`, []any{groupID, iamPermissionRequester}, iamPermissionInvitationID("b6", marker),
		groupID, iamPermissionRequester, nil, "iam.permission_request.cancel",
		"PENDING", "CANCELLED", 2)
	if result != groupID+"|CANCELLED|2" {
		t.Fatalf("permission request cancel = %q", result)
	}
}

func expireIAMPermissionRequest(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	marker string,
) {
	t.Helper()
	const groupID = "ae000000-0000-4000-8000-000000000001"
	result := runIAMPermissionTransition(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|',group_id::text,group_state,group_revision::text)
FROM iam_permission_request_expire_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,1,'PENDING'
)`, []any{groupID}, iamPermissionInvitationID("be", marker), groupID,
		iamPermissionReviewer, nil, "iam.permission_request.expire",
		"PENDING", "EXPIRED", 2)
	if result != groupID+"|EXPIRED|2" {
		t.Fatalf("permission request expire = %q", result)
	}
}

func reviewIAMPermissionRequest(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
	groupID string,
	marker string,
) {
	t.Helper()
	taskID := iamPermissionInvitationID("a4", marker)
	result := runIAMPermissionTransition(t, ctx, database, schema, `
SELECT pg_catalog.concat_ws('|',group_id::text,group_state,group_revision::text)
FROM iam_permission_request_review_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,1,'PENDING',$2,1,$3
)`, []any{groupID, iamPermissionReviewer, taskID},
		iamPermissionInvitationID("b4", marker), groupID, iamPermissionReviewer,
		taskID, "iam.permission_request.review", "PENDING", "IN_REVIEW", 2)
	if result != groupID+"|IN_REVIEW|2" {
		t.Fatalf("permission request review = %q", result)
	}
}

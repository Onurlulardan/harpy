//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"
)

const (
	iamRBACHTTPCreateOperationID = "b1000000-0000-4000-8000-000000000001"
	iamRBACHTTPUpdateOperationID = "b1000000-0000-4000-8000-000000000002"
	iamRBACHTTPRetireOperationID = "b1000000-0000-4000-8000-000000000003"
	iamRBACHTTPUserGroupID       = "b2000000-0000-4000-8000-000000000001"
)

func assertIAMRBACHTTPUserLifecycle(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	create := mustCreateIAMRBACHTTPAssignment(
		t, ctx, database, "t_iam_a", iamRBACHTTPCreateOperationID,
		iamRBACHTTPUserGroupID, "USER", iamUserTwoID,
		iamRBACReadActionID, "b3000000-0000-4000-8000-000000000001",
		"b4000000-0000-4000-8000-000000000001", 0xa1,
	)
	if create != iamRBACHTTPUserGroupID+"|1|1|false|true" {
		t.Fatalf("assignment create = %s", create)
	}
	assertIAMRBACHTTPReplayAndConflict(t, ctx, database)
	second := mustCreateIAMRBACHTTPAssignment(
		t, ctx, database, "t_iam_b", iamRBACHTTPCreateOperationID,
		iamRBACHTTPUserGroupID, "USER", iamUserTwoID,
		iamRBACReadActionID, "b3000000-0000-4000-8000-000000000001",
		"b4000000-0000-4000-8000-000000000001", 0xa1,
	)
	if second != iamRBACHTTPUserGroupID+"|1|1|false|true" {
		t.Fatalf("second tenant assignment = %s", second)
	}
	update := mustUpdateIAMRBACHTTPAssignment(
		t, ctx, database, "t_iam_a", iamRBACHTTPUpdateOperationID,
		iamRBACHTTPUserGroupID, "USER", iamUserTwoID,
		iamRBACWriteActionID, "b3000000-0000-4000-8000-000000000002",
		"b4000000-0000-4000-8000-000000000002", 1,
	)
	if update != iamRBACHTTPUserGroupID+"|2|2|false|true" {
		t.Fatalf("assignment update = %s", update)
	}
	assertIAMRBACHTTPBindingReplacement(t, ctx, database)
	retired := mustRetireIAMRBACHTTPAssignment(
		t, ctx, database, "t_iam_a", iamRBACHTTPRetireOperationID,
		iamRBACHTTPUserGroupID, 2,
	)
	if retired != iamRBACHTTPUserGroupID+"|3|3|false|true" {
		t.Fatalf("assignment retire = %s", retired)
	}
	assertIAMRBACHTTPRetiredState(t, ctx, database)
}

func assertIAMRBACHTTPReplayAndConflict(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	items := fmt.Sprintf(
		`[{"ordinal":1,"action_id":"%s","permission_id":"%s","binding_id":"%s"}]`,
		iamRBACReadActionID, "b3000000-0000-4000-8000-000000000001",
		"b4000000-0000-4000-8000-000000000001",
	)
	purpose := sha256.Sum256([]byte(
		"harpy/iam/rbac-step-up/v1|iam.rbac.assignment.create|",
	))
	arguments := []any{
		iamRBACHTTPCreateOperationID, bytes.Repeat([]byte{0xa1}, 32),
		iamRBACHTTPUserGroupID, "USER", iamUserTwoID, items, int64(1), int64(1),
		purpose[:],
	}
	replay := mustRunIAMAtomicText(
		t, ctx, database, "t_iam_a", replayIAMRBACHTTPAssignment, arguments...,
	)
	if replay != iamRBACHTTPUserGroupID+"|1|1|true" {
		t.Fatalf("assignment replay = %s", replay)
	}
	arguments[1] = bytes.Repeat([]byte{0xee}, 32)
	if _, err := runIAMAtomicText(
		ctx, database, "t_iam_a", replayIAMRBACHTTPAssignment, arguments...,
	); err == nil {
		t.Fatal("assignment operation/hash conflict was accepted")
	}
	if clock := iamRBACClock(t, ctx, database, "t_iam_a"); clock != 1 {
		t.Fatalf("assignment replay/conflict clock = %d", clock)
	}
}

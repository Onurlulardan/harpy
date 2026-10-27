//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMUserRelationReplayConflict(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)
	input := newIAMUserRelationsInput(actor, iamRelationUpdateUserID,
		"e3000000-0000-4000-8000-000000000001", "iam.identity.user.update", 1, 0x61)
	input.replaceOrganizations = true
	input.organizationIDs = iamUserRelationUUIDs(iamRelationOrgOneID)
	input.membershipIDs = iamUserRelationUUIDs(
		"d4000000-0000-4000-8000-000000000001")
	input.replaceRoles = true
	input.roleIDs = iamUserRelationUUIDs(iamRelationDefaultRoleID)
	input.assignmentIDs = iamUserRelationUUIDs(
		"d5000000-0000-4000-8000-000000000001")

	first := executeIAMUserRelations(t, ctx, database, "t_iam_a", input)
	assertIAMUserRelationResult(t, first, iamRelationUpdateUserID, 2, false)
	replay := executeIAMUserRelations(t, ctx, database, "t_iam_a", input)
	assertIAMUserRelationResult(t, replay, iamRelationUpdateUserID, 2, true)
	if replay.clock != first.clock {
		t.Fatalf("IAM relation replay clock = %d, want %d", replay.clock, first.clock)
	}

	conflict := input
	conflict.requestSHA = append([]byte(nil), input.requestSHA...)
	conflict.requestSHA[0] = 0x62
	assertIAMIdentitySQLState(t,
		rejectIAMUserRelations(t, ctx, database, "t_iam_a", conflict), "23505")
	assertIAMUserRelationState(t, readIAMUserRelationsState(
		t, ctx, database, "t_iam_a", iamRelationUpdateUserID,
	), 2, 1, iamRelationOrgOneID, iamRelationDefaultRoleID)
	if count := countIAMUserRelationReceipt(t, ctx, database, "t_iam_a",
		input.operationID); count != 1 {
		t.Fatalf("IAM relation replay receipt count = %d, want 1", count)
	}
}

func assertIAMUserRelationStatementRollback(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)
	input := newIAMUserRelationsInput(actor, iamRelationRollbackUserID,
		"e3000000-0000-4000-8000-000000000002", "iam.identity.user.update", 1, 0x63)
	input.replaceOrganizations = true
	input.organizationIDs = iamUserRelationUUIDs(iamRelationOrgTwoID)
	input.membershipIDs = iamUserRelationUUIDs(
		"d4000000-0000-4000-8000-000000000002")
	input.replaceRoles = true
	input.roleIDs = iamUserRelationUUIDs(iamRelationDefaultRoleID)
	input.assignmentIDs = iamUserRelationUUIDs(iamRBACProtectedAssignmentID)
	assertIAMIdentitySQLState(t,
		rejectIAMUserRelations(t, ctx, database, "t_iam_a", input), "23505")
	assertIAMUserRelationState(t, readIAMUserRelationsState(
		t, ctx, database, "t_iam_a", iamRelationRollbackUserID,
	), 1, 0, "", "")
	if count := countIAMUserRelationReceipt(t, ctx, database, "t_iam_a",
		input.operationID); count != 0 {
		t.Fatalf("rolled-back IAM relation receipt count = %d, want 0", count)
	}
}

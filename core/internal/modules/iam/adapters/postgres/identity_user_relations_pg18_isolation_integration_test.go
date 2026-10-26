//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMUserRelationTenantIsolation(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	actorA := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)
	input := newIAMUserRelationsInput(actorA, iamRelationIsolationUserID,
		"e4000000-0000-4000-8000-000000000001", "iam.identity.user.update", 1, 0x71)
	input.replaceOrganizations = true
	input.organizationIDs = iamUserRelationUUIDs(iamRelationOrgOneID)
	input.membershipIDs = iamUserRelationUUIDs(
		"d6000000-0000-4000-8000-000000000001")
	input.replaceRoles = true
	input.roleIDs = iamUserRelationUUIDs(iamRelationDefaultRoleID)
	input.assignmentIDs = iamUserRelationUUIDs(
		"d7000000-0000-4000-8000-000000000001")
	result := executeIAMUserRelations(t, ctx, database, "t_iam_a", input)
	assertIAMUserRelationResult(t, result, iamRelationIsolationUserID, 2, false)
	assertIAMUserRelationState(t, readIAMUserRelationsState(
		t, ctx, database, "t_iam_a", iamRelationIsolationUserID,
	), 2, 1, iamRelationOrgOneID, iamRelationDefaultRoleID)
	assertIAMUserRelationState(t, readIAMUserRelationsState(
		t, ctx, database, "t_iam_b", iamRelationIsolationUserID,
	), 1, 0, "", "")

	actorB := readIAMIdentityActor(t, ctx, database, "t_iam_b", iamIdentityActorID)
	crossScope := newIAMUserRelationsInput(actorB, iamRelationIsolationUserID,
		"e4000000-0000-4000-8000-000000000002", "iam.identity.user.update", 1, 0x72)
	crossScope.scope = "t_iam_a"
	crossScope.replaceOrganizations = true
	crossScope.organizationIDs = iamUserRelationUUIDs(iamRelationOrgTwoID)
	crossScope.membershipIDs = iamUserRelationUUIDs(
		"d6000000-0000-4000-8000-000000000002")
	assertIAMIdentitySQLState(t,
		rejectIAMUserRelations(t, ctx, database, "t_iam_b", crossScope), "42501")
	assertIAMUserRelationState(t, readIAMUserRelationsState(
		t, ctx, database, "t_iam_b", iamRelationIsolationUserID,
	), 1, 0, "", "")
}

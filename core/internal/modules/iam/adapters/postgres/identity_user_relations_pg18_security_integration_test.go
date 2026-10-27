//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMUserRelationSecurity(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)

	owner := newIAMUserRelationsInput(actor, iamUserTwoID,
		"e2000000-0000-4000-8000-000000000001", "iam.identity.user.update", 2, 0x51)
	owner.replaceOrganizations = true
	owner.organizationIDs = iamUserRelationUUIDs()
	owner.membershipIDs = iamUserRelationUUIDs()
	assertIAMIdentitySQLState(t,
		rejectIAMUserRelations(t, ctx, database, "t_iam_a", owner), "23514")

	protectedAdd := newIAMUserRelationsInput(actor, iamRelationSecurityUserID,
		"e2000000-0000-4000-8000-000000000002", "iam.identity.user.update", 1, 0x52)
	protectedAdd.replaceRoles = true
	protectedAdd.roleIDs = iamUserRelationUUIDs(iamRBACProtectedRoleID)
	protectedAdd.assignmentIDs = iamUserRelationUUIDs(
		"d3000000-0000-4000-8000-000000000001")
	assertIAMIdentitySQLState(t,
		rejectIAMUserRelations(t, ctx, database, "t_iam_a", protectedAdd), "42501")

	protectedRemove := newIAMUserRelationsInput(actor, iamUserTwoID,
		"e2000000-0000-4000-8000-000000000003", "iam.identity.user.update", 2, 0x53)
	protectedRemove.replaceRoles = true
	protectedRemove.roleIDs = iamUserRelationUUIDs()
	protectedRemove.assignmentIDs = iamUserRelationUUIDs()
	assertIAMIdentitySQLState(t,
		rejectIAMUserRelations(t, ctx, database, "t_iam_a", protectedRemove), "42501")

	self := newIAMUserRelationsInput(actor, iamIdentityActorID,
		"e2000000-0000-4000-8000-000000000004", "iam.identity.user.update",
		actor.revision, 0x54)
	self.replaceOrganizations = true
	self.organizationIDs = iamUserRelationUUIDs(iamRelationOrgOneID)
	self.membershipIDs = iamUserRelationUUIDs(
		"d3000000-0000-4000-8000-000000000002")
	assertIAMIdentitySQLState(t,
		rejectIAMUserRelations(t, ctx, database, "t_iam_a", self), "42501")

	stale := newIAMUserRelationsInput(actor, iamRelationSecurityUserID,
		"e2000000-0000-4000-8000-000000000005", "iam.identity.user.update", 2, 0x55)
	stale.replaceOrganizations = true
	stale.organizationIDs = iamUserRelationUUIDs(iamRelationOrgOneID)
	stale.membershipIDs = iamUserRelationUUIDs(
		"d3000000-0000-4000-8000-000000000003")
	assertIAMIdentitySQLState(t,
		rejectIAMUserRelations(t, ctx, database, "t_iam_a", stale), "40001")

	assertIAMUserRelationState(t, readIAMUserRelationsState(
		t, ctx, database, "t_iam_a", iamRelationSecurityUserID,
	), 1, 0, "", "")
}

//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
)

func assertIAMUserRelationCreateUpdate(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	actor := readIAMIdentityActor(t, ctx, database, "t_iam_a", iamIdentityActorID)
	create := newIAMUserRelationsInput(actor, iamRelationCreateUserID,
		"e1000000-0000-4000-8000-000000000001", "iam.identity.user.create", 1, 0x41)
	create.replaceOrganizations = true
	create.organizationIDs = iamUserRelationUUIDs(iamRelationOrgOneID)
	create.membershipIDs = iamUserRelationUUIDs(
		"d1000000-0000-4000-8000-000000000001")
	create.replaceRoles = true
	create.roleIDs = iamUserRelationUUIDs(iamRelationDefaultRoleID, iamRelationOrgRoleID)
	create.assignmentIDs = iamUserRelationUUIDs(
		"d2000000-0000-4000-8000-000000000001",
		"d2000000-0000-4000-8000-000000000002")
	created := executeIAMUserRelations(t, ctx, database, "t_iam_a", create)
	assertIAMUserRelationResult(t, created, iamRelationCreateUserID, 2, false)
	assertIAMUserRelationState(t, readIAMUserRelationsState(
		t, ctx, database, "t_iam_a", iamRelationCreateUserID,
	), 2, 1, iamRelationOrgOneID,
		iamRelationDefaultRoleID+","+iamRelationOrgRoleID)

	update := newIAMUserRelationsInput(actor, iamRelationCreateUserID,
		"e1000000-0000-4000-8000-000000000002", "iam.identity.user.update", 2, 0x42)
	update.replaceOrganizations = true
	update.organizationIDs = iamUserRelationUUIDs(iamRelationOrgTwoID)
	update.membershipIDs = iamUserRelationUUIDs(
		"d1000000-0000-4000-8000-000000000002")
	update.replaceRoles = true
	update.roleIDs = iamUserRelationUUIDs(iamRelationDefaultRoleID)
	update.assignmentIDs = iamUserRelationUUIDs(
		"d2000000-0000-4000-8000-000000000003")
	updated := executeIAMUserRelations(t, ctx, database, "t_iam_a", update)
	assertIAMUserRelationResult(t, updated, iamRelationCreateUserID, 3, false)
	assertIAMUserRelationState(t, readIAMUserRelationsState(
		t, ctx, database, "t_iam_a", iamRelationCreateUserID,
	), 3, 2, iamRelationOrgTwoID, iamRelationDefaultRoleID)

	noChange := update
	noChange.operationID = "e1000000-0000-4000-8000-000000000003"
	noChange.requestSHA[0] = 0x43
	noChange.expectedRevision = 3
	unchanged := executeIAMUserRelations(t, ctx, database, "t_iam_a", noChange)
	assertIAMUserRelationResult(t, unchanged, iamRelationCreateUserID, 4, false)
	assertIAMUserRelationState(t, readIAMUserRelationsState(
		t, ctx, database, "t_iam_a", iamRelationCreateUserID,
	), 4, 2, iamRelationOrgTwoID, iamRelationDefaultRoleID)

	profile := newIAMUserRelationsInput(actor, iamRelationCreateUserID,
		"e1000000-0000-4000-8000-000000000004", "iam.identity.user.update", 4, 0x44)
	profile.patchProfile = true
	profile.profileSchemaID = iamUserProfileSchemaID
	profile.profileDocument = []byte(`{}`)
	digest := sha256.Sum256(profile.profileDocument)
	profile.profileSHA = digest[:]
	profiled := executeIAMUserRelations(t, ctx, database, "t_iam_a", profile)
	assertIAMUserRelationResult(t, profiled, iamRelationCreateUserID, 5, false)
	assertIAMUserRelationState(t, readIAMUserRelationsState(
		t, ctx, database, "t_iam_a", iamRelationCreateUserID,
	), 5, 2, iamRelationOrgTwoID, iamRelationDefaultRoleID)
}

func assertIAMUserRelationResult(
	t *testing.T, result iamUserRelationsResult, id string, revision int64, replayed bool,
) {
	t.Helper()
	if result.id != id || result.revision != revision || result.clock < 1 ||
		result.replayed != replayed {
		t.Fatalf("IAM user relation result = %+v, want %s/%d/replay=%t",
			result, id, revision, replayed)
	}
}

func assertIAMUserRelationState(
	t *testing.T, state iamUserRelationsState, revision, authorization int64,
	organizations, roles string,
) {
	t.Helper()
	if state.revision != revision || state.authorizationVersion != authorization ||
		state.organizations != organizations || state.roles != roles {
		t.Fatalf("IAM user relation state = %+v, want rev=%d auth=%d org=%q roles=%q",
			state, revision, authorization, organizations, roles)
	}
}

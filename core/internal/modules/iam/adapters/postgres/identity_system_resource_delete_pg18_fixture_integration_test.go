//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"
)

type iamIdentityDeleteFixture struct {
	prefix, organization, role, user, identifier, accessMode string
	defaultAssignment, membership, organizationAssignment    string
	email, membershipKind                                    string
}

func newIAMIdentityDeleteFixture(prefix, email, kind string) iamIdentityDeleteFixture {
	return iamIdentityDeleteFixture{
		prefix: prefix, email: email, membershipKind: kind,
		organization:           prefix + "0000-0000-4000-8000-000000000001",
		role:                   prefix + "0000-0000-4000-8000-000000000002",
		user:                   prefix + "0000-0000-4000-8000-000000000003",
		identifier:             prefix + "0000-0000-4000-8000-000000000004",
		accessMode:             prefix + "0000-0000-4000-8000-000000000005",
		defaultAssignment:      prefix + "0000-0000-4000-8000-000000000006",
		membership:             prefix + "0000-0000-4000-8000-000000000007",
		organizationAssignment: prefix + "0000-0000-4000-8000-000000000008",
	}
}

func seedIAMIdentityDeleteGraph(
	t *testing.T, ctx context.Context, database *sql.DB,
	fixture iamIdentityDeleteFixture,
) {
	t.Helper()
	tenant := iamFoundationTenants[0]
	actor := readIAMIdentityActor(t, ctx, database, tenant.schema, iamIdentityActorID)
	profile := []byte(`{}`)
	profileSHA := sha256.Sum256(profile)
	emailSHA := sha256.Sum256([]byte(fixture.email))
	withIAMIdentityRuntimeTx(t, ctx, database, tenant.schema, func(tx *sql.Tx) {
		createUser := func() {
			callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteUserCreateSQL,
				tenant.id, fixture.prefix+"1000-0000-4000-8000-000000000003",
				bytes.Repeat([]byte{0x73}, 32), fixture.user, fixture.identifier,
				fixture.accessMode, fixture.defaultAssignment, iamUserProfileSchemaID,
				profile, profileSHA[:], fixture.email, fixture.email, emailSHA[:],
				"delete.fixture", bytes.Repeat([]byte{0x74}, 32), iamIdentityActorID,
				iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion)
		}
		if fixture.membershipKind == "owner" {
			createUser()
			callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteOwnedOrganizationCreateSQL,
				fixture.prefix+"1000-0000-4000-8000-000000000001",
				bytes.Repeat([]byte{0x71}, 32), fixture.organization,
				"Disposable "+fixture.prefix, "disposable-"+fixture.prefix,
				fixture.user, fixture.membership, iamOrganizationProfileSchemaID,
				profile, profileSHA[:], iamIdentityActorID, iamIdentityPermissionID,
				actor.tokenVersion, actor.authorizationVersion)
		} else {
			callIAMIdentityMutation(t, ctx, tx, iamIdentityOrganizationCreate,
				fixture.prefix+"1000-0000-4000-8000-000000000001",
				bytes.Repeat([]byte{0x71}, 32), fixture.organization,
				"Disposable "+fixture.prefix, "disposable-"+fixture.prefix,
				iamOrganizationProfileSchemaID, profile, profileSHA[:], iamIdentityActorID,
				iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion)
		}
		callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteRoleCreateSQL,
			tenant.id, fixture.prefix+"1000-0000-4000-8000-000000000002",
			bytes.Repeat([]byte{0x72}, 32), fixture.role, "organization",
			fixture.organization, fixture.organization, "Disposable Role "+fixture.prefix,
			"", false, iamRoleProfileSchemaID, profile, profileSHA[:],
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion)
		if fixture.membershipKind != "owner" {
			createUser()
			callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteMembershipAddSQL,
				fixture.prefix+"1000-0000-4000-8000-000000000004",
				bytes.Repeat([]byte{0x75}, 32), fixture.membership, fixture.organization,
				fixture.user, fixture.membershipKind, iamIdentityActorID,
				iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion)
		}
		callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteAssignmentGrantSQL,
			fixture.prefix+"1000-0000-4000-8000-000000000005",
			bytes.Repeat([]byte{0x76}, 32), fixture.organizationAssignment,
			fixture.user, fixture.role, "organization", fixture.organization,
			fixture.organization, fixture.membership, iamIdentityActorID,
			iamIdentityPermissionID, actor.tokenVersion, actor.authorizationVersion)
		callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteUserDisableSQL,
			fixture.prefix+"1000-0000-4000-8000-000000000006",
			bytes.Repeat([]byte{0x77}, 32), fixture.user, int64(3),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion)
	})
}

func callIAMIdentityMutation(
	t *testing.T, ctx context.Context, tx *sql.Tx,
	statement string, arguments ...any,
) (int64, bool) {
	t.Helper()
	var id string
	var revision, clock int64
	var replayed bool
	if err := tx.QueryRowContext(ctx, statement, arguments...).
		Scan(&id, &revision, &clock, &replayed); err != nil {
		t.Fatal(err)
	}
	if id == "" || revision < 1 || clock < 1 {
		t.Fatalf("invalid identity mutation result = %s/%d/%d/%t",
			id, revision, clock, replayed)
	}
	return revision, replayed
}

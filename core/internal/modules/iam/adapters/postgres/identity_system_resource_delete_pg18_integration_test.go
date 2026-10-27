//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func assertIAMIdentitySystemResourceDeleteCleanup(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	fixture := newIAMIdentityDeleteFixture(
		"b700", "phase14-cleanup@example.invalid", "member",
	)
	seedIAMIdentityDeleteGraph(t, ctx, database, fixture)
	tenant := iamFoundationTenants[0]
	actor := readIAMIdentityActor(t, ctx, database, tenant.schema, iamIdentityActorID)
	withIAMIdentityRuntimeTx(t, ctx, database, tenant.schema, func(tx *sql.Tx) {
		revokeIAMIdentityDeleteAssignments(t, ctx, tx, fixture, actor)
		callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteMembershipDeleteSQL,
			fixture.prefix+"2000-0000-4000-8000-000000000003",
			bytes.Repeat([]byte{0x83}, 32), fixture.membership, int64(1),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion)
		revision, replayed := callIAMIdentityMutation(
			t, ctx, tx, iamIdentityDeleteUserDeleteSQL,
			fixture.prefix+"2000-0000-4000-8000-000000000004",
			bytes.Repeat([]byte{0x84}, 32), fixture.user, int64(7),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion,
		)
		if revision != 8 || replayed {
			t.Fatalf("user delete = %d/%t", revision, replayed)
		}
		callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteRoleRetireSQL,
			fixture.prefix+"2000-0000-4000-8000-000000000005",
			bytes.Repeat([]byte{0x85}, 32), fixture.role, int64(1),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion)
		callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteOrganizationDisableSQL,
			fixture.prefix+"2000-0000-4000-8000-000000000006",
			bytes.Repeat([]byte{0x86}, 32), fixture.organization, int64(1),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion)
		callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteOrganizationDeleteSQL,
			fixture.prefix+"2000-0000-4000-8000-000000000007",
			bytes.Repeat([]byte{0x87}, 32), fixture.organization, int64(2),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion)
	})
	assertIAMIdentityDeleteTerminalState(t, ctx, database, tenant.schema, fixture)
	withIAMIdentityRuntimeTx(t, ctx, database, tenant.schema, func(tx *sql.Tx) {
		organizationRevision, organizationReplayed := callIAMIdentityMutation(
			t, ctx, tx, iamIdentityDeleteOrganizationDeleteSQL,
			fixture.prefix+"2000-0000-4000-8000-000000000007",
			bytes.Repeat([]byte{0x87}, 32), fixture.organization, int64(2),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion,
		)
		if organizationRevision != 3 || !organizationReplayed {
			t.Fatalf("organization delete replay = %d/%t",
				organizationRevision, organizationReplayed)
		}
		revision, replayed := callIAMIdentityMutation(
			t, ctx, tx, iamIdentityDeleteUserDeleteSQL,
			fixture.prefix+"2000-0000-4000-8000-000000000004",
			bytes.Repeat([]byte{0x84}, 32), fixture.user, int64(4),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion,
		)
		if revision != 8 || !replayed {
			t.Fatalf("user delete replay = %d/%t", revision, replayed)
		}
	})
	probeIAMIdentityRuntimeTx(t, ctx, database, "t_iam_b", func(tx *sql.Tx) {
		row := tx.QueryRowContext(ctx, iamIdentityDeleteOrganizationDeleteSQL,
			fixture.prefix+"3000-0000-4000-8000-000000000001",
			bytes.Repeat([]byte{0x88}, 32), fixture.organization, int64(2),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion)
		var id string
		var revision, clock int64
		var replayed bool
		assertIAMIdentitySQLState(
			t, row.Scan(&id, &revision, &clock, &replayed), "40001",
		)
	})
}

func revokeIAMIdentityDeleteAssignments(
	t *testing.T, ctx context.Context, tx *sql.Tx,
	fixture iamIdentityDeleteFixture, actor iamIdentityActor,
) {
	t.Helper()
	for index, assignment := range []string{
		fixture.defaultAssignment, fixture.organizationAssignment,
	} {
		callIAMIdentityMutation(t, ctx, tx, iamIdentityDeleteAssignmentRevokeSQL,
			fixture.prefix+"2000-0000-4000-8000-00000000000"+string(rune('1'+index)),
			bytes.Repeat([]byte{byte(0x81 + index)}, 32), assignment, int64(1),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion)
	}
}

func assertIAMIdentityDeleteTerminalState(
	t *testing.T, ctx context.Context, database *sql.DB,
	schema string, fixture iamIdentityDeleteFixture,
) {
	t.Helper()
	connection, transaction := beginIAMTenantTx(t, ctx, database, schema)
	defer connection.Close()
	defer transaction.Rollback()
	var user, role, organization, membership string
	var nonterminal int
	err := transaction.QueryRowContext(ctx, `
SELECT usr.status,role.status,organization.status,membership.state,
 (SELECT count(*) FROM iam_user_role_assignments
  WHERE user_id=$1 AND state<>'revoked')
FROM iam_users usr JOIN iam_roles role ON role.id=$2
JOIN iam_organizations organization ON organization.id=$3
JOIN iam_organization_memberships membership ON membership.id=$4
WHERE usr.id=$1`, fixture.user, fixture.role, fixture.organization, fixture.membership).
		Scan(&user, &role, &organization, &membership, &nonterminal)
	if err != nil || user != "tombstoned" || role != "retired" ||
		organization != "tombstoned" || membership != "tombstoned" ||
		nonterminal != 0 {
		t.Fatalf("terminal cleanup = %s/%s/%s/%s/%d: %v",
			user, role, organization, membership, nonterminal, err)
	}
}

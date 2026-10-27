//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
)

func assertIAMIdentityUserDeleteOwnerRollback(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	fixture := newIAMIdentityDeleteFixture(
		"b800", "phase14-owner@example.invalid", "owner",
	)
	seedIAMIdentityDeleteGraph(t, ctx, database, fixture)
	tenant := iamFoundationTenants[0]
	actor := readIAMIdentityActor(t, ctx, database, tenant.schema, iamIdentityActorID)
	probeIAMIdentityRuntimeTx(t, ctx, database, tenant.schema, func(tx *sql.Tx) {
		revokeIAMIdentityDeleteAssignments(t, ctx, tx, fixture, actor)
		row := tx.QueryRowContext(ctx, iamIdentityDeleteMembershipDeleteSQL,
			fixture.prefix+"2000-0000-4000-8000-000000000003",
			bytes.Repeat([]byte{0x93}, 32), fixture.membership, int64(1),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion)
		var id string
		var revision, clock int64
		var replayed bool
		assertIAMIdentitySQLState(
			t, row.Scan(&id, &revision, &clock, &replayed), "23514",
		)
	})
	probeIAMIdentityRuntimeTx(t, ctx, database, tenant.schema, func(tx *sql.Tx) {
		revokeIAMIdentityDeleteAssignments(t, ctx, tx, fixture, actor)
		row := tx.QueryRowContext(ctx, iamIdentityDeleteUserDeleteSQL,
			fixture.prefix+"2000-0000-4000-8000-000000000004",
			bytes.Repeat([]byte{0x94}, 32), fixture.user, int64(6),
			iamIdentityActorID, iamIdentityPermissionID,
			actor.tokenVersion, actor.authorizationVersion)
		var id string
		var revision, clock int64
		var replayed bool
		assertIAMIdentitySQLState(
			t, row.Scan(&id, &revision, &clock, &replayed), "23514",
		)
	})
	connection, transaction := beginIAMTenantTx(
		t, ctx, database, tenant.schema,
	)
	defer connection.Close()
	defer transaction.Rollback()
	var user, membership, defaultAssignment, organizationAssignment string
	err := transaction.QueryRowContext(ctx, `
SELECT usr.status,membership.state,default_assignment.state,organization_assignment.state
FROM iam_users usr
JOIN iam_organization_memberships membership ON membership.id=$2
JOIN iam_user_role_assignments default_assignment ON default_assignment.id=$3
JOIN iam_user_role_assignments organization_assignment ON organization_assignment.id=$4
WHERE usr.id=$1`, fixture.user, fixture.membership, fixture.defaultAssignment,
		fixture.organizationAssignment).Scan(
		&user, &membership, &defaultAssignment, &organizationAssignment,
	)
	if err != nil || user != "disabled" || membership != "active" ||
		defaultAssignment != "active" || organizationAssignment != "active" {
		t.Fatalf("owner rollback = %s/%s/%s/%s: %v",
			user, membership, defaultAssignment, organizationAssignment, err)
	}
}

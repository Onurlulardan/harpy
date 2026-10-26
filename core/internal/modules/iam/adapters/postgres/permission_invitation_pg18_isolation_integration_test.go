//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
)

const (
	iamIsolationInvitation = "d1000000-0000-4000-8000-000000000001"
	iamIsolationToken      = "d2000000-0000-4000-8000-000000000001"
)

func assertIAMPermissionInvitationTenantIsolation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	createIAMIsolationInvitation(
		t, ctx, database, iamFoundationTenants[0], "isolation-a@example.com",
	)
	createIAMIsolationInvitation(
		t, ctx, database, iamFoundationTenants[1], "isolation-b@example.com",
	)
	revokeIAMInvitation(
		t, ctx, database, iamFoundationTenants[0].schema, iamIsolationInvitation,
	)
	state := mustRunIAMPGFirstText(t, ctx, database, iamFoundationTenants[1].schema, `
SELECT invitation_state FROM iam_invitation_read_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2
)`, iamInvitationManager, iamIsolationInvitation)
	if state != "pending" {
		t.Fatalf("tenant B invitation state changed through tenant A: %s", state)
	}
	assertIAMPermissionInvitationScopeSubstitution(t, ctx, database)
	assertIAMPermissionInvitationACL(t, ctx, database)
}

func createIAMIsolationInvitation(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	tenant iamFoundationTenant,
	email string,
) {
	t.Helper()
	result := mustRunIAMPGFirstText(t, ctx, database, tenant.schema, `
SELECT pg_catalog.concat_ws('|',invitation_id::text,invitation_state)
FROM iam_invitation_create_v1(
    (pg_catalog.current_schemas(false))[2]::name,
    $1,$2,$2,$3,$4,'tenant',$5,NULL,$6,1,$7,$8,'one-time-1',$9,$10
)`, iamIsolationInvitation, email, []byte(strings.Repeat("z", 32)),
		iamInvitationRole, tenant.id, iamInvitationManager, iamIsolationToken,
		[]byte("isolation-token1"), []byte(strings.Repeat("i", 32)),
		[]byte(strings.Repeat("j", 32)))
	if result != iamIsolationInvitation+"|pending" {
		t.Fatalf("create isolation invitation in %s = %q", tenant.schema, result)
	}
}

func assertIAMPermissionInvitationScopeSubstitution(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	err := rejectIAMPGFirst(t, ctx, database, "t_iam_a", `
SELECT * FROM iam_invitation_read_v1($1::name,$2,$3)
`, "t_iam_b", iamInvitationManager, iamIsolationInvitation)
	assertIAMSQLState(t, err, "42501")
	err = rejectIAMPGFirst(t, ctx, database, "t_iam_a", `
SELECT * FROM iam_permission_request_read_v1($1::name,$2,$3)
`, "t_iam_b", iamPermissionReviewer,
		iamPermissionInvitationID("a1", "8a"))
	assertIAMSQLState(t, err, "42501")
}

func assertIAMPermissionInvitationACL(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	err := rejectIAMPGFirst(t, ctx, database, "t_iam_a", `
UPDATE iam_invitations SET updated_at=updated_at WHERE id=$1
`, iamIsolationInvitation)
	assertIAMSQLState(t, err, "42501")
	err = rejectIAMPGFirst(t, ctx, database, "t_iam_a", `
SELECT iam_invitation_manager_v1($1)
`, iamInvitationManager)
	assertIAMSQLState(t, err, "42501")
	err = rejectIAMPGFirst(t, ctx, database, "t_iam_a", `
SELECT iam_user_has_permission_v1($1,$2)
`, iamPermissionReviewer, iamPermissionReview)
	assertIAMSQLState(t, err, "42501")
}

//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
)

func assertIAMInvitationRegistrationV2(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant, marker string,
) {
	t.Helper()
	fixture := newIAMInvitationRegistrationFixture(tenant, marker)
	createIAMInvitationRegistrationV2(t, ctx, database, tenant, fixture)
	registration, directory := invitationRegistrationArguments(fixture)
	assertInvitationDirectoryRollback(
		t, ctx, database, tenant, fixture, registration, directory,
	)
	expectedClock := invitationRegistrationClock(
		t, ctx, database, tenant.schema,
	) + 1
	assertInvitationRegistrationCommit(
		t, ctx, database, tenant, fixture, registration, directory,
		expectedClock, false,
	)
	assertInvitationRegistrationCommit(
		t, ctx, database, tenant, fixture, registration, directory,
		expectedClock, true,
	)
	assertInvitationRegistrationState(
		t, ctx, database, tenant, fixture, "accepted", "consumed", 0, 1,
	)
}

func assertInvitationDirectoryRollback(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant, fixture iamInvitationRegistrationFixture,
	registration, directory []any,
) {
	t.Helper()
	connection, transaction, err := beginIAMPGFirstRuntimeTx(
		ctx, database, tenant.schema,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var result string
	if err := transaction.QueryRowContext(
		ctx, iamInvitationRegisterSQL, registration...,
	).Scan(&result); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("registration rollback setup: %s", iamPGErrorDetail(err))
	}
	badDirectory := append([]any(nil), directory...)
	badDirectory[3] = []byte{1}
	arguments := append([]any{tenant.id}, badDirectory...)
	_, err = transaction.ExecContext(ctx, iamInvitationDirectorySQL, arguments...)
	assertIAMSQLState(t, err, "42501")
	if err := transaction.Rollback(); err != nil {
		t.Fatalf("rollback failed directory projection: %v", err)
	}
	assertInvitationRegistrationState(
		t, ctx, database, tenant, fixture, "pending", "pending", 1, 0,
	)
}

func assertInvitationRegistrationCommit(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant, fixture iamInvitationRegistrationFixture,
	registration, directory []any, expectedClock int64, replay bool,
) {
	t.Helper()
	connection, transaction, err := beginIAMPGFirstRuntimeTx(
		ctx, database, tenant.schema,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	var result string
	if err := transaction.QueryRowContext(
		ctx, iamInvitationRegisterSQL, registration...,
	).Scan(&result); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("registration commit: %s", iamPGErrorDetail(err))
	}
	want := fmt.Sprintf("iam.invitation.accepted|2|%d|false", expectedClock)
	if replay {
		want = fmt.Sprintf("iam.invitation.accepted|2|%d|true", expectedClock)
	}
	if result != want {
		_ = transaction.Rollback()
		t.Fatalf("registration result=%q want=%q", result, want)
	}
	arguments := append([]any{tenant.id}, directory...)
	var revision int64
	if err := transaction.QueryRowContext(
		ctx, iamInvitationDirectorySQL, arguments...,
	).Scan(&revision); err != nil {
		_ = transaction.Rollback()
		t.Fatalf("directory projection: %s", iamPGErrorDetail(err))
	}
	if revision != 1 {
		_ = transaction.Rollback()
		t.Fatalf("directory source revision=%d", revision)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit registration: %v", err)
	}
}

func invitationRegistrationClock(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) int64 {
	t.Helper()
	var value int64
	mustQueryIAMTenantValue(
		t, ctx, database, schema,
		"SELECT clock_value FROM iam_authorization_clock WHERE singleton", &value,
	)
	return value
}

func assertInvitationRegistrationState(
	t *testing.T, ctx context.Context, database *sql.DB,
	tenant iamFoundationTenant, fixture iamInvitationRegistrationFixture,
	wantInvitation, wantToken string, wantArtifact, wantUser int64,
) {
	t.Helper()
	query := `SELECT invitation.state,token.state,
 (SELECT pg_catalog.count(*) FROM iam_invitation_delivery_artifacts
  WHERE invitation_id='` + fixture.invitationID + `'),
 (SELECT pg_catalog.count(*) FROM iam_users WHERE id='` + fixture.userID + `')
FROM iam_invitations invitation JOIN iam_one_time_tokens token
 ON token.id='` + fixture.tokenID + `' AND token.invitation_id=invitation.id
WHERE invitation.id='` + fixture.invitationID + `'`
	var invitationState, tokenState string
	var artifactCount, userCount int64
	mustQueryIAMTenantValue(
		t, ctx, database, tenant.schema, query,
		&invitationState, &tokenState, &artifactCount, &userCount,
	)
	if invitationState != wantInvitation || tokenState != wantToken ||
		artifactCount != wantArtifact || userCount != wantUser {
		t.Fatalf("registration state=%s/%s artifact=%d user=%d",
			invitationState, tokenState, artifactCount, userCount)
	}
	var directoryCount int64
	if err := database.QueryRowContext(ctx, `SELECT pg_catalog.count(*)
FROM platform.iam_login_directory WHERE tenant_id=$1 AND user_id=$2`,
		tenant.id, fixture.userID).Scan(&directoryCount); err != nil {
		t.Fatal(err)
	}
	if directoryCount != wantUser {
		t.Fatalf("directory rows=%d want=%d", directoryCount, wantUser)
	}
}

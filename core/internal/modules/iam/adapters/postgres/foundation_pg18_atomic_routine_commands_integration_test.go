//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMAtomicRoutineCommands(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
	schema string,
) {
	t.Helper()
	assertIAMOneTimeCrashAndReplay(t, ctx, database, schema)
	assertIAMInvitationAcceptCAS(t, ctx, database, schema)
	assertIAMChallengeResolveCAS(t, ctx, database, schema)
	assertIAMRecoveryConsumeCAS(t, ctx, database, schema)
	assertIAMPortalSetupCAS(t, ctx, database, schema)
}

func assertIAMOneTimeCrashAndReplay(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	const call = `SELECT iam_one_time_consume_v1(
    $1, '88000000-0000-4000-8000-000000000002', 1,
    'email_verification_v1'
)`
	connection, transaction, err := beginIAMAtomicRuntimeTx(ctx, database, schema)
	if err != nil {
		t.Fatal(err)
	}
	var outcome string
	if err := transaction.QueryRowContext(ctx, call, schema).Scan(&outcome); err != nil ||
		outcome != "consumed" {
		_ = transaction.Rollback()
		_ = connection.Close()
		t.Fatalf("%s IAM crash candidate = %q: %v", schema, outcome, err)
	}
	_ = transaction.Rollback()
	_ = connection.Close()
	var state string
	var revision int64
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT state, revision FROM iam_one_time_tokens
WHERE id = '88000000-0000-4000-8000-000000000002'`, &state, &revision)
	if state != "pending" || revision != 1 {
		t.Fatalf("%s IAM crash rollback state/revision = %s/%d", schema, state, revision)
	}
	if outcome = mustRunIAMAtomicText(t, ctx, database, schema, call, schema); outcome != "consumed" {
		t.Fatalf("%s IAM one-time outcome = %s", schema, outcome)
	}
	_, err = runIAMAtomicText(ctx, database, schema, call, schema)
	assertIAMSQLState(t, err, "40001")
}

func assertIAMInvitationAcceptCAS(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	const call = `SELECT iam_invitation_accept_v1(
    $1, '97000000-0000-4000-8000-000000000001', 1,
    '98000000-0000-4000-8000-000000000001', 1,
    '51000000-0000-4000-8000-000000000003',
    '96000000-0000-4000-8000-000000000001'
)`
	if outcome := mustRunIAMAtomicText(t, ctx, database, schema, call, schema); outcome != "accepted" {
		t.Fatalf("%s IAM invitation outcome = %s", schema, outcome)
	}
	_, err := runIAMAtomicText(ctx, database, schema, call, schema)
	assertIAMSQLState(t, err, "40001")
	var invitationState, tokenState string
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT invitation.state, token.state
FROM iam_invitations AS invitation
JOIN iam_one_time_tokens AS token ON token.invitation_id = invitation.id
WHERE invitation.id = '97000000-0000-4000-8000-000000000001'`,
		&invitationState, &tokenState)
	if invitationState != "accepted" || tokenState != "consumed" {
		t.Fatalf("%s IAM invitation/token = %s/%s", schema, invitationState, tokenState)
	}
}

func assertIAMChallengeResolveCAS(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	const call = `SELECT iam_challenge_resolve_v1(
    $1, '99000000-0000-4000-8000-000000000001', $2, $3, $4
)`
	_, err := runIAMAtomicText(ctx, database, schema, call, schema, 1, 1, "reject")
	assertIAMSQLState(t, err, "40001")
	if outcome := mustRunIAMAtomicText(
		t, ctx, database, schema, call, schema, 1, 0, "reject",
	); outcome != "pending" {
		t.Fatalf("%s IAM challenge reject outcome = %s", schema, outcome)
	}
	if outcome := mustRunIAMAtomicText(
		t, ctx, database, schema, call, schema, 2, 1, "consume",
	); outcome != "consumed" {
		t.Fatalf("%s IAM challenge consume outcome = %s", schema, outcome)
	}
	_, err = runIAMAtomicText(ctx, database, schema, call, schema, 2, 1, "consume")
	assertIAMSQLState(t, err, "40001")
}

func assertIAMRecoveryConsumeCAS(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	const call = `SELECT iam_recovery_consume_v1(
    $1, '85000000-0000-4000-8000-000000000001', 1
)`
	if version := mustRunIAMAtomicBigint(t, ctx, database, schema, call, schema); version != 2 {
		t.Fatalf("%s IAM recovery token version = %d, want 2", schema, version)
	}
	_, err := runIAMAtomicText(ctx, database, schema, call, schema)
	assertIAMSQLState(t, err, "40001")
	var state, authenticatorState string
	mustQueryIAMTenantValue(t, ctx, database, schema, `
SELECT code.state, authenticator.state
FROM iam_recovery_codes AS code
JOIN iam_authenticators AS authenticator ON authenticator.id = code.authenticator_id
WHERE code.id = '85000000-0000-4000-8000-000000000001'`, &state, &authenticatorState)
	if state != "consumed" || authenticatorState != "renewal_required" {
		t.Fatalf("%s IAM recovery/authenticator = %s/%s", schema, state, authenticatorState)
	}
}

func assertIAMPortalSetupCAS(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	const call = `SELECT iam_portal_setup_complete_v1(
    $1, '53000000-0000-4000-8000-000000000003', 1,
    '98000000-0000-4000-8000-000000000004', 1
)`
	if outcome := mustRunIAMAtomicText(t, ctx, database, schema, call, schema); outcome != "completed" {
		t.Fatalf("%s IAM portal setup outcome = %s", schema, outcome)
	}
	_, err := runIAMAtomicText(ctx, database, schema, call, schema)
	assertIAMSQLState(t, err, "40001")
}

//go:build integration

package postgres

import (
	"bytes"
	"context"
	"database/sql"
	"testing"
	"time"
)

type iamLoginState struct {
	userRevision, tokenVersion, authorizationVersion int64
	accessModeID                                     string
	accessRevision                                   int64
	credentialVersion, credentialRevision            int64
}

func assertIAMAuthLoginAndSession(
	t *testing.T,
	ctx context.Context,
	database iamPlatformGuardDatabase,
	fixture iamAuthRuntimeFixture,
) {
	t.Helper()
	transaction, err := beginIAMAuthRuntimeTx(ctx, database.runtime, fixture.tenant.schema)
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := resolveIAMLoginEvidence(ctx, transaction, fixture)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	state := readIAMLoginState(t, ctx, transaction, evidence, fixture)
	var used iamLoginState
	err = transaction.QueryRowContext(ctx, `
SELECT user_revision,token_version,authorization_version,
       ''::text,0::bigint,credential_version,credential_revision
FROM iam_password_use_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,
 false,NULL,NULL,NULL,NULL,NULL)`,
		evidence.userID, state.userRevision, state.tokenVersion,
		state.authorizationVersion, state.credentialVersion,
		state.credentialRevision,
	).Scan(&used.userRevision, &used.tokenVersion, &used.authorizationVersion,
		&used.accessModeID, &used.accessRevision,
		&used.credentialVersion, &used.credentialRevision)
	if err != nil || used.credentialRevision != state.credentialRevision+1 {
		_ = transaction.Rollback()
		t.Fatalf("%s IAM password use: %+v %v", fixture.tenant.schema, used, err)
	}
	issuedAt, idle, absolute, retain := issueIAMRuntimeSession(
		t, ctx, transaction, fixture, used, state,
	)
	_, err = transaction.ExecContext(ctx, iamCreateLocatorSQL,
		fixture.selector, fixture.tenant.id, iamAuthTokenID, iamAuthFamilyID,
		int64(1), int64(1), int64(1), fixture.locatorSHA, "active",
		issuedAt, issuedAt, absolute, retain)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatalf("%s create issued locator: %v", fixture.tenant.schema, err)
	}
	if !idle.Equal(issuedAt.Add(30*24*time.Hour)) ||
		!absolute.Equal(issuedAt.Add(90*24*time.Hour)) ||
		!retain.Equal(absolute.Add(24*time.Hour)) {
		_ = transaction.Rollback()
		t.Fatal("IAM session TTL contract drifted")
	}
	if err := transaction.Commit(); err != nil {
		t.Fatalf("commit %s IAM login: %v", fixture.tenant.schema, err)
	}
}

func readIAMLoginState(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	evidence iamLoginEvidence,
	fixture iamAuthRuntimeFixture,
) iamLoginState {
	t.Helper()
	var state iamLoginState
	var tag []byte
	err := transaction.QueryRowContext(ctx, `
SELECT user_revision,token_version,authorization_version,access_mode_id,
       access_mode_revision,credential_version,credential_revision,tag
FROM iam_login_credential_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,'email_v1',$3,$4,
 'operator_console')`, evidence.userID, evidence.identifierID,
		evidence.revision, evidence.sourceSHA256,
	).Scan(&state.userRevision, &state.tokenVersion, &state.authorizationVersion,
		&state.accessModeID, &state.accessRevision, &state.credentialVersion,
		&state.credentialRevision, &tag)
	if err != nil || !bytes.Equal(tag, fixture.credentialTag) {
		t.Fatalf("%s IAM login credential: tag=%x err=%v",
			fixture.tenant.schema, tag, err)
	}
	return state
}

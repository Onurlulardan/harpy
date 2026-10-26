//go:build integration

package postgres

import (
	"bytes"
	"context"
	"testing"
)

func assertIAMPasswordReauthenticationState(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	proof iamPasswordReauthentication,
) {
	t.Helper()
	var challengeState, oldState, successorState, acr string
	var amrPassword, amrWebAuthn, amrTOTP, amrRecovery bool
	var purpose []byte
	var passwordRevision int64
	queryIAMAuthState(t, ctx, database.migrator, proof.fixture.tenant.schema, `
SELECT challenge.state,old_token.state,new_token.state,new_token.acr,
       new_token.amr_password,new_token.amr_webauthn,new_token.amr_totp,
       new_token.amr_recovery,new_token.step_up_purpose_sha256,
       password.revision
FROM iam_auth_challenges AS challenge
JOIN iam_refresh_tokens AS old_token
  ON old_token.id='64000000-0000-4000-8000-000000000001'::uuid
JOIN iam_refresh_tokens AS new_token
  ON new_token.id='64000000-0000-4000-8000-000000000010'::uuid
JOIN iam_password_credentials AS password
  ON password.user_id='51000000-0000-4000-8000-000000000001'::uuid
WHERE challenge.id='65000000-0000-4000-8000-000000000001'::uuid`,
		&challengeState, &oldState,
		&successorState, &acr, &amrPassword, &amrWebAuthn, &amrTOTP,
		&amrRecovery, &purpose, &passwordRevision)
	if challengeState != "consumed" || oldState != "consumed" ||
		successorState != "active" || acr != "step_up" || !amrPassword ||
		amrWebAuthn || amrTOTP || amrRecovery ||
		!bytes.Equal(purpose, proof.purpose) ||
		passwordRevision != proof.passwordRevision+1 {
		t.Fatal("password reauthentication state contract drifted")
	}
}

func assertIAMPasswordReauthenticationReplayRejected(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	proof iamPasswordReauthentication,
) {
	t.Helper()
	transaction, err := database.runtime.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	bindIAMRuntimeSchema(t, ctx, transaction, proof.fixture.tenant.schema)
	if _, err := completeIAMPasswordReauthentication(ctx, transaction, proof); err == nil {
		t.Fatal("consumed reauthentication challenge was replayed")
	}
}

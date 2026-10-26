//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

func assertIAMSessionRevocation(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	proof iamPasswordReauthentication,
) {
	t.Helper()
	outcome := revokeIAMSecuritySession(
		t, ctx, database, proof.successorSelector, true,
	)
	if outcome != "revoked" {
		t.Fatalf("session revocation outcome=%s", outcome)
	}
	assertIAMRevokedSessionState(t, ctx, database, proof)
	outcome = revokeIAMSecuritySession(
		t, ctx, database, proof.successorSelector, false,
	)
	if outcome != "terminal" {
		t.Fatalf("idempotent session revocation outcome=%s", outcome)
	}
}

func revokeIAMSecuritySession(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	selector []byte, project bool,
) string {
	t.Helper()
	transaction, err := database.runtime.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	evidence, err := resolveIAMRefreshEvidence(ctx, transaction, selector)
	if err != nil {
		t.Fatal(err)
	}
	schema, err := iamAuthSchemaForTenant(evidence.tenantID)
	if err != nil {
		t.Fatal(err)
	}
	bindIAMRuntimeSchema(t, ctx, transaction, schema)
	state, err := readIAMRefreshState(ctx, transaction, evidence)
	if err != nil {
		t.Fatal(err)
	}
	outcome, projection, err := executeIAMSessionRevocation(
		ctx, transaction, evidence, state,
	)
	if err != nil {
		t.Fatal(err)
	}
	if project {
		if err := advanceIAMRefreshProjection(
			ctx, transaction, evidence, projection,
		); err != nil {
			t.Fatal(err)
		}
	} else if len(projection.locatorSelector) != 0 ||
		projection.locatorRevision.Valid {
		t.Fatal("terminal session revocation emitted a projection")
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return outcome
}

func executeIAMSessionRevocation(
	ctx context.Context, transaction *sql.Tx,
	evidence iamRefreshEvidence, state iamRefreshState,
) (string, iamRotationProjection, error) {
	var outcome string
	var projection iamRotationProjection
	err := transaction.QueryRowContext(ctx, `
SELECT outcome,locator_selector,locator_expected_revision,locator_revision,
       locator_source_sha256,locator_state,locator_projected_at,
       locator_retain_until
FROM iam_session_revoke_v1(
 (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4,$5,$6,$7,$8)`,
		evidence.tokenID, state.familyID, state.tokenRevision,
		state.familyRevision, state.sessionID, state.sessionRevision,
		state.userID, state.userRevision,
	).Scan(
		&outcome, &projection.locatorSelector,
		&projection.locatorExpected, &projection.locatorRevision,
		&projection.locatorSHA, &projection.locatorState,
		&projection.locatorProjected, &projection.locatorRetain,
	)
	return outcome, projection, err
}

func assertIAMRevokedSessionState(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	proof iamPasswordReauthentication,
) {
	t.Helper()
	var tokenState, familyState, sessionState, reason, locatorState string
	var tokenRevision, familyRevision, sessionRevision, locatorRevision int64
	queryIAMAuthState(t, ctx, database.migrator, proof.fixture.tenant.schema, `
SELECT token.state,token.revision,family.state,family.revision,
       family.revoke_reason,session.state,session.revision
FROM iam_refresh_tokens AS token
JOIN iam_refresh_families AS family ON family.id=token.family_id
JOIN iam_sessions AS session ON session.id=family.session_id
WHERE token.id='64000000-0000-4000-8000-000000000010'::uuid`,
		&tokenState, &tokenRevision, &familyState, &familyRevision,
		&reason, &sessionState, &sessionRevision)
	if tokenState != "revoked" || tokenRevision != 2 ||
		familyState != "revoked" ||
		familyRevision != proof.state.familyRevision+2 || reason != "logout" ||
		sessionState != "revoked" ||
		sessionRevision != proof.state.sessionRevision+2 {
		t.Fatal("session revocation state contract drifted")
	}
	if err := database.migrator.QueryRowContext(ctx, `
SELECT source_token_state,source_revision
FROM platform.iam_refresh_locators WHERE selector=$1`,
		proof.successorSelector,
	).Scan(&locatorState, &locatorRevision); err != nil ||
		locatorState != "revoked" || locatorRevision != 2 {
		t.Fatalf("revoked locator state=%s revision=%d err=%v",
			locatorState, locatorRevision, err)
	}
}

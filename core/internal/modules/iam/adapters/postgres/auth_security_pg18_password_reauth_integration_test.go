//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
	"github.com/google/uuid"
)

const (
	iamReauthenticationChallengeID = "65000000-0000-4000-8000-000000000001"
	iamReauthenticationSuccessorID = "64000000-0000-4000-8000-000000000010"
)

type iamPasswordReauthentication struct {
	fixture                           iamAuthRuntimeFixture
	evidence                          iamRefreshEvidence
	state                             iamRefreshState
	binding                           []byte
	purpose                           []byte
	origin                            []byte
	contextSHA                        []byte
	challengeHMAC                     []byte
	passwordVersion, passwordRevision int64
	successorSelector, successorHMAC  []byte
}

func assertIAMPasswordReauthentication(
	t *testing.T, ctx context.Context, database iamPlatformGuardDatabase,
	fixture iamAuthRuntimeFixture,
) iamPasswordReauthentication {
	t.Helper()
	transaction, err := database.runtime.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	evidence, err := resolveIAMRefreshEvidence(ctx, transaction, fixture.selector)
	if err != nil {
		t.Fatal(err)
	}
	bindIAMRuntimeSchema(t, ctx, transaction, fixture.tenant.schema)
	state, err := readIAMRefreshState(ctx, transaction, evidence)
	if err != nil {
		t.Fatal(err)
	}
	proof := newIAMPasswordReauthentication(t, fixture, evidence, state)
	startIAMPasswordReauthentication(t, ctx, transaction, proof)
	readIAMPasswordReauthentication(t, ctx, transaction, &proof)
	projection, err := completeIAMPasswordReauthentication(ctx, transaction, proof)
	if err != nil {
		t.Fatalf("%s complete password reauthentication: %v", fixture.tenant.schema, err)
	}
	if projection.outcome != "rotated" {
		t.Fatalf("reauthentication outcome=%s", projection.outcome)
	}
	if err := advanceIAMRefreshProjection(ctx, transaction, evidence, projection); err != nil {
		t.Fatal(err)
	}
	if err := createIAMSuccessorProjection(
		ctx, transaction, evidence, state.familyID,
		iamReauthenticationSuccessorID, projection,
	); err != nil {
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	assertIAMPasswordReauthenticationState(t, ctx, database, proof)
	return proof
}

func newIAMPasswordReauthentication(
	t *testing.T, fixture iamAuthRuntimeFixture,
	evidence iamRefreshEvidence, state iamRefreshState,
) iamPasswordReauthentication {
	t.Helper()
	binding, err := authentication.ReauthenticationBindingSHA256(
		uuid.MustParse(fixture.tenant.id), uuid.MustParse(state.userID),
		uuid.MustParse(state.sessionID), state.tokenVersion, state.authorization,
		authentication.AccessModeOperator, uuid.Nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	purpose := sha256.Sum256([]byte("iam.security.change-password"))
	origin := sha256.Sum256([]byte("https://app.example.com"))
	contextSHA, err := authentication.ReauthenticationContextSHA256(
		uuid.MustParse(fixture.tenant.id), uuid.MustParse(state.userID),
		uuid.MustParse(state.sessionID), purpose, 1,
		authentication.ReauthenticationPassword, origin,
	)
	if err != nil {
		t.Fatal(err)
	}
	selectorSHA := sha256.Sum256([]byte(
		"iam.reauthentication.successor.selector:" + fixture.tenant.id,
	))
	successorHMAC := sha256.Sum256([]byte(
		"iam.reauthentication.successor.hmac:" + fixture.tenant.id,
	))
	return iamPasswordReauthentication{
		fixture: fixture, evidence: evidence, state: state,
		binding: binding[:], purpose: purpose[:], origin: origin[:],
		contextSHA: contextSHA[:], challengeHMAC: bytes.Repeat([]byte{0xa1}, 32),
		successorSelector: append([]byte(nil), selectorSHA[:16]...),
		successorHMAC:     append([]byte(nil), successorHMAC[:]...),
	}
}

func bindIAMRuntimeSchema(
	t *testing.T, ctx context.Context, transaction *sql.Tx, schema string,
) {
	t.Helper()
	want := "pg_catalog, " + schema
	if _, err := transaction.ExecContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)", want); err != nil {
		t.Fatal(err)
	}
}

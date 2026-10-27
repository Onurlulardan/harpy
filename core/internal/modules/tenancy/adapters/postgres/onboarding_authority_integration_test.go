//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSaaSOnboardingAuthorityCASFencesActivationPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := onboardingTestDatabase(t, ctx, "saas")
	fixture := newOnboardingFixture(t, "saas")
	attemptID := uuid.New()

	transaction := beginRuntimeTransaction(t, ctx, database)
	acceptOnboarding(t, ctx, transaction, fixture, fixture.manifestSHA)
	revision, _ := claimOnboarding(t, ctx, transaction, fixture, attemptID, 1, 60)
	_, revision = recordMigration(t, ctx, transaction, fixture, attemptID, revision)
	if err := transaction.Commit(); err != nil || revision != 3 {
		t.Fatalf("SaaS migration revision=%d err=%v", revision, err)
	}

	transaction = beginRuntimeTransaction(t, ctx, database)
	if err := authorityAckError(ctx, transaction, fixture, strings.Repeat("0", 64), 3); err == nil {
		t.Fatal("authority ack with drifted state digest was accepted")
	}
	_ = transaction.Rollback()
	state, revision := onboardingState(t, ctx, database, fixture.operationID)
	if state != "authority_pending" || revision != 3 {
		t.Fatalf("failed CAS mutated operation=%s/%d", state, revision)
	}

	ackSHA := strings.Repeat("7", 64)
	transaction = beginRuntimeTransaction(t, ctx, database)
	state, revision = authorityAck(t, ctx, transaction, fixture, ackSHA, 3)
	if err := transaction.Commit(); err != nil || state != "authority_acknowledged" || revision != 4 {
		t.Fatalf("authority CAS=%s/%d err=%v", state, revision, err)
	}

	completionReceiptID := uuid.New()
	transaction = beginRuntimeTransaction(t, ctx, database)
	revision, returnedAck := activateOnboarding(t, ctx, transaction, fixture, 4)
	if revision != 5 || returnedAck != ackSHA {
		t.Fatalf("SaaS activation revision=%d ack=%q", revision, returnedAck)
	}
	completeOnboarding(t, ctx, transaction, fixture, completionReceiptID, revision)
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	state, revision = onboardingState(t, ctx, database, fixture.operationID)
	if state != "succeeded" || revision != 6 {
		t.Fatalf("SaaS completion=%s/%d", state, revision)
	}
}

func authorityAck(
	t *testing.T,
	ctx context.Context,
	transaction *sql.Tx,
	fixture onboardingFixture,
	ackSHA string,
	expectedRevision int64,
) (string, int64) {
	t.Helper()
	var operationID uuid.UUID
	var state string
	var revision int64
	var updatedAt time.Time
	err := authorityAckRow(ctx, transaction, fixture, fixture.stateSHA, ackSHA,
		expectedRevision).Scan(&operationID, &state, &revision, &updatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if operationID != fixture.operationID || updatedAt.IsZero() {
		t.Fatalf("authority ack identity=%s updated=%s", operationID, updatedAt)
	}
	return state, revision
}

func authorityAckError(
	ctx context.Context,
	transaction *sql.Tx,
	fixture onboardingFixture,
	stateSHA string,
	expectedRevision int64,
) error {
	return authorityAckRow(ctx, transaction, fixture, stateSHA, strings.Repeat("7", 64),
		expectedRevision).Scan(new(uuid.UUID), new(string), new(int64), new(time.Time))
}

func authorityAckRow(
	ctx context.Context,
	transaction *sql.Tx,
	fixture onboardingFixture,
	stateSHA, ackSHA string,
	expectedRevision int64,
) *sql.Row {
	return transaction.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_onboarding_authority_ack_v1($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		fixture.operationID, fixture.registryBindingID, "environment", fixture.environmentID,
		int64(1), nil, stateSHA, fixture.proposalSHA, ackSHA, int64(1),
		fixture.createdAt, expectedRevision)
}

//go:build integration

package postgres

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOnboardingOperationReplayAndAtomicActivationPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := onboardingTestDatabase(t, ctx, "onpremise")
	fixture := newOnboardingFixture(t, "onpremise")

	transaction := beginRuntimeTransaction(t, ctx, database)
	state, revision := acceptOnboarding(t, ctx, transaction, fixture, fixture.manifestSHA)
	if state != "accepted" || revision != 1 {
		t.Fatalf("initial transition=%s/%d", state, revision)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}

	transaction = beginRuntimeTransaction(t, ctx, database)
	state, revision = acceptOnboarding(t, ctx, transaction, fixture, fixture.manifestSHA)
	if state != "accepted" || revision != 1 {
		t.Fatalf("exact replay=%s/%d", state, revision)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	var stepCount int
	if err := database.QueryRowContext(ctx, `SELECT count(*) FROM
platform.tenancy_onboarding_steps WHERE operation_id=$1`, fixture.operationID).
		Scan(&stepCount); err != nil || stepCount != 9 {
		t.Fatalf("step plan count=%d err=%v", stepCount, err)
	}

	transaction = beginRuntimeTransaction(t, ctx, database)
	arguments := fixture.acceptArguments(strings.Repeat("f", 64))
	err := transaction.QueryRowContext(ctx, onboardingAcceptSQL, arguments...).Scan(
		new(uuid.UUID), new(string), new(int64), new(time.Time))
	if err == nil {
		t.Fatal("different manifest replay was accepted")
	}
	_ = transaction.Rollback()

	if _, err := database.ExecContext(ctx, `UPDATE platform.tenancy_onboarding_operations
SET tenant_id=$2 WHERE operation_id=$1`, fixture.operationID, uuid.New()); err == nil {
		t.Fatal("immutable operation plan was mutated")
	}
	transaction = beginRuntimeTransaction(t, ctx, database)
	if _, err := transaction.ExecContext(ctx, `UPDATE platform.tenancy_onboarding_operations
SET state=state WHERE operation_id=$1`, fixture.operationID); err == nil {
		t.Fatal("runtime received direct onboarding table DML")
	}
	_ = transaction.Rollback()

	attemptID := uuid.New()
	transaction = beginRuntimeTransaction(t, ctx, database)
	revision, _ = claimOnboarding(t, ctx, transaction, fixture, attemptID, 1, 60)
	if revision != 2 {
		t.Fatalf("claim revision=%d", revision)
	}
	_, revision = recordMigration(t, ctx, transaction, fixture, attemptID, revision)
	if revision != 3 {
		t.Fatalf("migration revision=%d", revision)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}

	rolledBackReceiptID := uuid.New()
	transaction = beginRuntimeTransaction(t, ctx, database)
	revision, ackSHA := activateOnboarding(t, ctx, transaction, fixture, 3)
	if revision != 4 || ackSHA != "" {
		t.Fatalf("onprem activation claim revision=%d ack=%q", revision, ackSHA)
	}
	completeOnboarding(t, ctx, transaction, fixture, rolledBackReceiptID, revision)
	if err := transaction.Rollback(); err != nil {
		t.Fatal(err)
	}
	state, revision = onboardingState(t, ctx, database, fixture.operationID)
	if state != "migration_ready" || revision != 3 {
		t.Fatalf("rollback leaked state=%s/%d", state, revision)
	}
	var rolledBackReceipts int
	if err := database.QueryRowContext(ctx, `SELECT count(*) FROM
platform.tenancy_onboarding_receipts WHERE receipt_id=$1`, rolledBackReceiptID).
		Scan(&rolledBackReceipts); err != nil || rolledBackReceipts != 0 {
		t.Fatalf("rollback receipt count=%d err=%v", rolledBackReceipts, err)
	}

	completionReceiptID := uuid.New()
	transaction = beginRuntimeTransaction(t, ctx, database)
	revision, _ = activateOnboarding(t, ctx, transaction, fixture, 3)
	completeOnboarding(t, ctx, transaction, fixture, completionReceiptID, revision)
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	state, revision = onboardingState(t, ctx, database, fixture.operationID)
	if state != "succeeded" || revision != 5 {
		t.Fatalf("commit state=%s/%d", state, revision)
	}

	transaction = beginRuntimeTransaction(t, ctx, database)
	completeOnboarding(t, ctx, transaction, fixture, completionReceiptID, 999)
	state, revision = acceptOnboarding(t, ctx, transaction, fixture, fixture.manifestSHA)
	if err := transaction.Commit(); err != nil || state != "succeeded" || revision != 5 {
		t.Fatalf("terminal replay=%s/%d err=%v", state, revision, err)
	}
}

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

func TestOnboardingLeaseExpiryAndAmbiguousReconcilePG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := onboardingTestDatabase(t, ctx, "onpremise")
	fixture := newOnboardingFixture(t, "onpremise")
	attemptID := uuid.New()

	transaction := beginRuntimeTransaction(t, ctx, database)
	acceptOnboarding(t, ctx, transaction, fixture, fixture.manifestSHA)
	revision, leaseUntil := claimOnboarding(t, ctx, transaction, fixture, attemptID, 1, 1)
	if revision != 2 || leaseUntil.IsZero() {
		t.Fatalf("short lease revision=%d until=%s", revision, leaseUntil)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	if _, err := database.ExecContext(ctx, "SELECT pg_catalog.pg_sleep(1.2)"); err != nil {
		t.Fatal(err)
	}

	receiptID := uuid.New()
	receiptSHA, evidenceSHA := strings.Repeat("d", 64), strings.Repeat("e", 64)
	transaction = beginRuntimeTransaction(t, ctx, database)
	state, revision := reconcileOnboarding(t, ctx, transaction, fixture.operationID,
		receiptID, receiptSHA, evidenceSHA, 2)
	if state != "accepted" || revision != 3 {
		t.Fatalf("reconcile transition=%s/%d", state, revision)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	var outcome string
	var leaseCount int
	if err := database.QueryRowContext(ctx, `SELECT outcome FROM
platform.tenancy_onboarding_attempts WHERE attempt_id=$1`, attemptID).Scan(&outcome); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRowContext(ctx, `SELECT count(*) FROM
platform.tenancy_onboarding_leases WHERE operation_id=$1`, fixture.operationID).
		Scan(&leaseCount); err != nil || outcome != "ambiguous" || leaseCount != 0 {
		t.Fatalf("expired attempt outcome=%s leases=%d err=%v", outcome, leaseCount, err)
	}

	transaction = beginRuntimeTransaction(t, ctx, database)
	state, revision = reconcileOnboarding(t, ctx, transaction, fixture.operationID,
		receiptID, receiptSHA, evidenceSHA, 999)
	if err := transaction.Commit(); err != nil || state != "accepted" || revision != 3 {
		t.Fatalf("reconcile replay=%s/%d err=%v", state, revision, err)
	}

	secondAttemptID := uuid.New()
	transaction = beginRuntimeTransaction(t, ctx, database)
	revision, _ = claimOnboarding(t, ctx, transaction, fixture, secondAttemptID, 3, 60)
	if err := transaction.Commit(); err != nil || revision != 4 {
		t.Fatalf("recovery claim revision=%d err=%v", revision, err)
	}
	var attemptNumber int64
	if err := database.QueryRowContext(ctx, `SELECT attempt_number FROM
platform.tenancy_onboarding_attempts WHERE attempt_id=$1`, secondAttemptID).
		Scan(&attemptNumber); err != nil || attemptNumber != 2 {
		t.Fatalf("recovery attempt number=%d err=%v", attemptNumber, err)
	}

	transaction = beginRuntimeTransaction(t, ctx, database)
	if err := reconcileOnboardingError(ctx, transaction, fixture.operationID,
		receiptID, strings.Repeat("f", 64), evidenceSHA, 4); err == nil {
		t.Fatal("conflicting reconcile receipt replay was accepted")
	}
	_ = transaction.Rollback()
}

func reconcileOnboarding(
	t *testing.T,
	ctx context.Context,
	transaction interface {
		QueryRowContext(context.Context, string, ...any) *sql.Row
	},
	operationID, receiptID uuid.UUID,
	receiptSHA, evidenceSHA string,
	expectedRevision int64,
) (string, int64) {
	t.Helper()
	var returnedID uuid.UUID
	var state string
	var revision int64
	var updatedAt time.Time
	err := transaction.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_onboarding_reconcile_v1($1,$2,$3,$4,$5)`, operationID, receiptID,
		receiptSHA, evidenceSHA, expectedRevision).Scan(&returnedID, &state, &revision, &updatedAt)
	if err != nil {
		t.Fatal(err)
	}
	return state, revision
}

func reconcileOnboardingError(
	ctx context.Context,
	transaction *sql.Tx,
	operationID, receiptID uuid.UUID,
	receiptSHA, evidenceSHA string,
	expectedRevision int64,
) error {
	return transaction.QueryRowContext(ctx, `SELECT * FROM
platform.tenancy_onboarding_reconcile_v1($1,$2,$3,$4,$5)`, operationID, receiptID,
		receiptSHA, evidenceSHA, expectedRevision).Scan(
		new(uuid.UUID), new(string), new(int64), new(time.Time))
}

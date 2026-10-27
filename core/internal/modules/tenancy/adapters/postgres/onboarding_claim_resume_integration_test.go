//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOnboardingClaimV2PreservesPostMigrationStateAfterLeaseExpiryPG18(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := onboardingTestDatabase(t, ctx, "onpremise")
	fixture := newOnboardingFixture(t, "onpremise")
	firstAttempt := uuid.New()

	transaction := beginRuntimeTransaction(t, ctx, database)
	acceptOnboarding(t, ctx, transaction, fixture, fixture.manifestSHA)
	revision, leaseUntil := claimOnboarding(t, ctx, transaction, fixture,
		firstAttempt, 1, 1)
	_, revision = recordMigration(t, ctx, transaction, fixture, firstAttempt, revision)
	if err := transaction.Commit(); err != nil || revision != 3 {
		t.Fatalf("commit migration-ready revision=%d err=%v", revision, err)
	}
	if _, err := database.ExecContext(ctx, `SELECT pg_catalog.pg_sleep(
GREATEST(0, EXTRACT(EPOCH FROM ($1::timestamptz - pg_catalog.clock_timestamp()))) + 0.05)`,
		leaseUntil); err != nil {
		t.Fatal(err)
	}

	secondAttempt := uuid.New()
	transaction = beginRuntimeTransaction(t, ctx, database)
	revision, resumedUntil := claimOnboarding(t, ctx, transaction, fixture, secondAttempt, 3, 1)
	if err := transaction.Commit(); err != nil || revision != 4 {
		t.Fatalf("resume claim revision=%d err=%v", revision, err)
	}
	state, storedRevision := onboardingState(t, ctx, database, fixture.operationID)
	if state != "migration_ready" || storedRevision != 4 {
		t.Fatalf("resume regressed state=%s revision=%d", state, storedRevision)
	}
	transaction = beginRuntimeTransaction(t, ctx, database)
	refreshedRevision, renewedUntil := claimOnboarding(
		t, ctx, transaction, fixture, secondAttempt, 4, 300,
	)
	if err := transaction.Commit(); err != nil || refreshedRevision != 4 ||
		!renewedUntil.After(resumedUntil.Add(4*time.Minute)) {
		t.Fatalf("lease renewal revision=%d until=%s err=%v",
			refreshedRevision, renewedUntil, err)
	}

	var firstOutcome string
	if err := database.QueryRowContext(ctx, `SELECT outcome FROM
platform.tenancy_onboarding_attempts WHERE attempt_id=$1`, firstAttempt).
		Scan(&firstOutcome); err != nil || firstOutcome != "expired" {
		t.Fatalf("first attempt outcome=%q err=%v", firstOutcome, err)
	}
}

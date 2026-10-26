//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"
)

type terminalGapRedactionResult struct {
	outbox int
	roots  int
	err    error
}

func scheduleTerminalGapEvidenceExpiry(
	t *testing.T,
	ctx context.Context,
	value terminalGapRetryCase,
	delay time.Duration,
) time.Time {
	t.Helper()
	ageOutboxForGapRedactionRace(
		t, ctx, value.fixture, value.sourceEventID.String(),
	)
	transaction, err := value.fixture.owner.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `
ALTER TABLE platform.dead_letters
DISABLE TRIGGER dead_letters_redaction_guard`); err != nil {
		t.Fatal(err)
	}
	var deadline time.Time
	if err = transaction.QueryRowContext(ctx, `
UPDATE platform.dead_letters
SET full_payload_expires_at=pg_catalog.clock_timestamp()+$2::interval
WHERE id=$1
RETURNING full_payload_expires_at`,
		value.rootID, delay.String(),
	).Scan(&deadline); err != nil {
		t.Fatal(err)
	}
	if _, err = transaction.ExecContext(ctx, `
ALTER TABLE platform.dead_letters
ENABLE ALWAYS TRIGGER dead_letters_redaction_guard`); err != nil {
		t.Fatal(err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
	return deadline
}

func waitForTerminalGapEvidenceExpiry(
	t *testing.T,
	ctx context.Context,
	value terminalGapRetryCase,
) {
	t.Helper()
	waitCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var expired bool
		err := value.fixture.owner.QueryRowContext(waitCtx, `
SELECT pg_catalog.clock_timestamp() >= full_payload_expires_at
FROM platform.dead_letters WHERE id=$1`, value.rootID).Scan(&expired)
		if err != nil {
			t.Fatalf("observe terminal gap expiry: %v", err)
		}
		if expired {
			return
		}
		select {
		case <-waitCtx.Done():
			t.Fatal("terminal gap evidence did not expire")
		case <-ticker.C:
		}
	}
}

func waitForTerminalGapRetryInsertBarrier(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	retry <-chan terminalGapRetryRaceResult,
) {
	t.Helper()
	waitCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		var blocked bool
		err := fixture.owner.QueryRowContext(waitCtx, `
SELECT EXISTS (
 SELECT 1 FROM pg_catalog.pg_locks AS waiter
 WHERE waiter.locktype='advisory' AND NOT waiter.granted
   AND waiter.database=(SELECT oid FROM pg_catalog.pg_database
                        WHERE datname=pg_catalog.current_database())
)`).Scan(&blocked)
		if err != nil {
			t.Fatalf("observe terminal gap retry barrier: %v", err)
		}
		if blocked {
			return
		}
		select {
		case result := <-retry:
			t.Fatalf(
				"terminal gap retry bypassed insert barrier: %v",
				result.err,
			)
		case <-waitCtx.Done():
			t.Fatal("terminal gap retry did not reach insert barrier")
		case <-ticker.C:
		}
	}
}

func runTerminalGapSafeRedaction(
	ctx context.Context,
	fixture *eventGapPGFixture,
	result chan<- terminalGapRedactionResult,
) {
	var value terminalGapRedactionResult
	counts, err := runRetentionCycleCounts(ctx, fixture.runtime)
	value.outbox, value.roots, value.err = int(counts[1]), int(counts[2]), err
	result <- value
}

func runTerminalGapRootRedaction(
	ctx context.Context,
	fixture *eventGapPGFixture,
	result chan<- terminalGapRedactionResult,
) {
	var value terminalGapRedactionResult
	counts, err := runRetentionCycleCounts(ctx, fixture.runtime)
	value.outbox, value.roots, value.err = int(counts[1]), int(counts[2]), err
	result <- value
}

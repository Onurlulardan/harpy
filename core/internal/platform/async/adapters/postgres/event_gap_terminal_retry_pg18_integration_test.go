//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestTerminalEventGapRetryEvidenceOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
			defer cancel()
			fixture := newEventGapPGFixture(t, ctx, origin)
			t.Run("expired_original_root", func(t *testing.T) {
				assertExpiredGapRootRetryRejected(t, ctx, &fixture)
			})
			t.Run("producer_only_generation_switch", func(t *testing.T) {
				assertStaleGapProducerRetryRejected(t, ctx, &fixture)
			})
		})
	}
}

func assertExpiredGapRootRetryRejected(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	value := newTerminalGapRetryCase(t, ctx, fixture)
	expireOriginalGapRoot(t, ctx, value)
	retryJobID := uuid.New()
	_, err := retryTerminalGap(ctx, value, retryJobID)
	if !errors.Is(err, ErrDeadLetterPayloadExpired) {
		t.Fatalf("expired original gap root retry error = %v", err)
	}
	assertRejectedGapRetryAtomic(t, ctx, value, retryJobID)
}

func assertStaleGapProducerRetryRejected(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	value := newTerminalGapRetryCase(t, ctx, fixture)
	groupID, generation := makeGapProducerGroupStale(t, ctx, value)
	defer setGapProducerGroupGeneration(
		t, ctx, value.fixture, groupID, generation,
	)
	retryJobID := uuid.New()
	_, err := retryTerminalGap(ctx, value, retryJobID)
	if !errors.Is(err, ErrDeadLetterAuthority) {
		t.Fatalf("stale gap producer retry error = %v", err)
	}
	assertRejectedGapRetryAtomic(t, ctx, value, retryJobID)
}

func expireOriginalGapRoot(
	t *testing.T,
	ctx context.Context,
	value terminalGapRetryCase,
) {
	t.Helper()
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
	if _, err = transaction.ExecContext(ctx, `
UPDATE platform.dead_letters
SET full_payload_expires_at=recorded_at+interval '1 microsecond'
WHERE id=$1`, value.rootID); err != nil {
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
}

func assertRejectedGapRetryAtomic(
	t *testing.T,
	ctx context.Context,
	value terminalGapRetryCase,
	retryJobID uuid.UUID,
) {
	t.Helper()
	var actions, jobs int
	err := value.fixture.runtime.QueryRow(ctx, `
SELECT
 (SELECT count(*) FROM platform.dead_letters
  WHERE root_id=$1 AND record_kind='retry_action'),
 (SELECT count(*) FROM platform.jobs WHERE id=$2)`,
		value.terminalRoot, retryJobID,
	).Scan(&actions, &jobs)
	if err != nil || actions != 0 || jobs != 0 {
		t.Fatalf("rejected gap retry actions/jobs = %d/%d: %v", actions, jobs, err)
	}
}

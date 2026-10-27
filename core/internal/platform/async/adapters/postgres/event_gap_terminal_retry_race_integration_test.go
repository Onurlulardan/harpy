//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	deadletter "github.com/Onurlulardan/harpy/core/internal/platform/async/deadletter"
	"github.com/google/uuid"
)

type terminalGapRetryRaceResult struct {
	result deadletter.RetryResult
	err    error
}

func TestTerminalEventGapRetrySerializesWithRedactionOnPostgreSQL184(
	t *testing.T,
) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			fixture := newEventGapPGFixture(t, ctx, origin)
			t.Run("redaction_wins", func(t *testing.T) {
				assertTerminalGapRedactionWinsRetry(t, ctx, &fixture)
			})
			t.Run("retry_wins", func(t *testing.T) {
				assertTerminalGapRetryWinsRedaction(t, ctx, &fixture)
			})
		})
	}
}

func assertTerminalGapRetryWinsRedaction(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	value := newTerminalGapRetryCase(t, ctx, fixture)
	barrier := installTerminalGapRetryInsertBarrier(t, ctx, value)
	defer barrier.cleanup()
	scheduleTerminalGapEvidenceExpiry(t, ctx, value, 5*time.Second)

	retryJobID := uuid.New()
	retry := make(chan terminalGapRetryRaceResult, 1)
	go func() {
		result, err := retryTerminalGap(ctx, value, retryJobID)
		retry <- terminalGapRetryRaceResult{result: result, err: err}
	}()
	waitForTerminalGapRetryInsertBarrier(t, ctx, fixture, retry)
	waitForTerminalGapEvidenceExpiry(t, ctx, value)

	redaction := make(chan terminalGapRedactionResult, 1)
	go runTerminalGapSafeRedaction(ctx, fixture, redaction)
	waitForGapRedactionLock(t, ctx, fixture)
	barrier.release()

	retried := receiveTerminalGapRetry(t, ctx, retry)
	if retried.err != nil || retried.result.ResultID != retryJobID {
		t.Fatalf("gap retry race result = %#v: %v", retried.result, retried.err)
	}
	redacted := receiveTerminalGapRedaction(t, ctx, redaction)
	if redacted.err != nil || redacted.outbox != 0 || redacted.roots != 0 {
		t.Fatalf("retry-wins redaction = %#v", redacted)
	}
	assertTerminalGapRetryEvidenceLive(t, ctx, value, retryJobID)
}

func assertTerminalGapRedactionWinsRetry(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	value := newTerminalGapRetryCase(t, ctx, fixture)
	expireOriginalGapRoot(t, ctx, value)
	barrier := installTerminalGapRedactionBarrier(t, ctx, value)
	defer barrier.cleanup()

	redaction := make(chan terminalGapRedactionResult, 1)
	go runTerminalGapRootRedaction(ctx, fixture, redaction)
	waitForGapParkBarrier(t, ctx, fixture)
	retryJobID := uuid.New()
	retry := make(chan terminalGapRetryRaceResult, 1)
	go func() {
		result, err := retryTerminalGap(ctx, value, retryJobID)
		retry <- terminalGapRetryRaceResult{result: result, err: err}
	}()
	waitForTerminalGapRetryRootLock(t, ctx, fixture)
	barrier.release()

	redacted := receiveTerminalGapRedaction(t, ctx, redaction)
	if redacted.err != nil || redacted.roots < 1 {
		t.Fatalf("redaction-wins result = %#v", redacted)
	}
	retried := receiveTerminalGapRetry(t, ctx, retry)
	if !errors.Is(retried.err, ErrDeadLetterPayloadExpired) {
		t.Fatalf("redaction-wins retry error = %v", retried.err)
	}
	assertRejectedGapRetryAtomic(t, ctx, value, retryJobID)
	assertTerminalGapRootRedacted(t, ctx, value)
}

func receiveTerminalGapRetry(
	t *testing.T,
	ctx context.Context,
	result <-chan terminalGapRetryRaceResult,
) terminalGapRetryRaceResult {
	t.Helper()
	select {
	case value := <-result:
		return value
	case <-ctx.Done():
		t.Fatalf("receive terminal gap retry: %v", ctx.Err())
		return terminalGapRetryRaceResult{}
	}
}

func receiveTerminalGapRedaction(
	t *testing.T,
	ctx context.Context,
	result <-chan terminalGapRedactionResult,
) terminalGapRedactionResult {
	t.Helper()
	select {
	case value := <-result:
		return value
	case <-ctx.Done():
		t.Fatalf("receive terminal gap redaction: %v", ctx.Err())
		return terminalGapRedactionResult{}
	}
}

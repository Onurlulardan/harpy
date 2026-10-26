//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func disposeStaleConcurrently(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	authority JobAuthority,
	command StaleJobDispositionCommand,
	stage string,
) StaleJobDispositionResult {
	t.Helper()
	type outcome struct {
		result StaleJobDispositionResult
		err    error
	}
	outcomes := make(chan outcome, 2)
	for range 2 {
		go func() {
			result, err := fixture.store.DisposeStaleJob(ctx, authority, command)
			outcomes <- outcome{result: result, err: err}
		}()
	}
	var created, replayed StaleJobDispositionResult
	creators := 0
	for range 2 {
		result := <-outcomes
		if result.err != nil {
			t.Fatalf("concurrent %s: %v", stage, result.err)
		}
		if result.result.Created() {
			creators++
			created = result.result
		} else {
			replayed = result.result
		}
	}
	if creators != 1 {
		t.Fatalf("%s creators = %d, want 1", stage, creators)
	}
	assertSameStaleDispositionResult(t, created, replayed, stage)
	return created
}

func assertStaleDispositionReplay(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	authority JobAuthority,
	command StaleJobDispositionCommand,
	original StaleJobDispositionResult,
	claim JobClaim,
	sourceID uuid.UUID,
) {
	t.Helper()
	assertSameStaleDispositionResult(t, original,
		disposeStaleAgain(t, ctx, fixture, authority, command), "claimed")
	cancellation, err := NewJobCancellation(
		"OPERATOR_CANCELLED", eventGapDigest("cancel:"+claim.JobID().String()),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fixture.store.CancelJob(ctx, claim, cancellation); err != nil {
		t.Fatalf("terminalize adopted job: %v", err)
	}
	assertSameStaleDispositionResult(t, original,
		disposeStaleAgain(t, ctx, fixture, authority, command), "terminal")
	redactStaleSourcePayloadAfterThirtyDays(t, ctx, fixture, sourceID)
	assertSameStaleDispositionResult(t, original,
		disposeStaleAgain(t, ctx, fixture, authority, command), "redacted source")
}

func disposeStaleAgain(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	authority JobAuthority,
	command StaleJobDispositionCommand,
) StaleJobDispositionResult {
	t.Helper()
	result, err := fixture.store.DisposeStaleJob(ctx, authority, command)
	if err != nil {
		t.Fatalf("replay stale disposition: %v", err)
	}
	return result
}

func assertSameStaleDispositionResult(
	t *testing.T,
	want StaleJobDispositionResult,
	got StaleJobDispositionResult,
	stage string,
) {
	t.Helper()
	wantJob, wantJobOK := want.ResultJobID()
	gotJob, gotJobOK := got.ResultJobID()
	wantRoot, wantRootOK := want.ResultRootID()
	gotRoot, gotRootOK := got.ResultRootID()
	if got.Created() || !got.CompletedAt().Equal(want.CompletedAt()) ||
		got.ActionID() != want.ActionID() || gotJobOK != wantJobOK ||
		gotJob != wantJob || gotRootOK != wantRootOK || gotRoot != wantRoot {
		t.Fatalf("%s replay drifted: want=%#v got=%#v", stage, want, got)
	}
}

func redactStaleSourcePayloadAfterThirtyDays(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	sourceID uuid.UUID,
) {
	t.Helper()
	transaction, err := fixture.owner.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx,
		"ALTER TABLE platform.jobs DISABLE TRIGGER jobs_transition_guard",
	); err != nil {
		t.Fatalf("disable stale source guard: %v", err)
	}
	if _, err := transaction.ExecContext(ctx, `
UPDATE platform.jobs
SET payload_ref=NULL,
    payload_redacted_at=completed_at+interval '31 days',
    updated_at=completed_at+interval '31 days'
WHERE id=$1`, sourceID); err != nil {
		t.Fatalf("redact stale source payload: %v", err)
	}
	if _, err := transaction.ExecContext(ctx,
		"ALTER TABLE platform.jobs ENABLE ALWAYS TRIGGER jobs_transition_guard",
	); err != nil {
		t.Fatalf("restore stale source guard: %v", err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

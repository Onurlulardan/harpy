//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestLiveEventGapEvidenceSurvivesRetentionOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
			defer cancel()
			fixture := newEventGapPGFixture(t, ctx, origin)
			assertLiveEventGapRetention(t, ctx, &fixture)
		})
	}
}

func assertLiveEventGapRetention(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	fixture.seedDelivered(t, ctx, 1)
	fixture.seedDeliveredAt(t, ctx, 2, time.Now().UTC().Add(-32*24*time.Hour))
	fixture.deliver(t, ctx, 2)
	rootID, jobID := fixture.gapIDs(t, ctx, 2)
	normalEventID := seedExpiredNormalOutbox(t, ctx, fixture)
	normalRootID := insertExpiredNormalDeadLetter(
		t, ctx, fixture, rootID, normalEventID,
	)
	ageGapRetentionRows(
		t, ctx, fixture, rootID,
		uuid.MustParse(fixture.envelopes[2].EventID), normalEventID,
	)
	assertRuntimeCannotRemoveGapEvidence(t, ctx, fixture, rootID)
	assertRetiredGapRedactionEntrypoints(t, ctx, fixture, jobID, "queued")
	outboxRedacted, deadRedacted := runGapSafeRedaction(t, ctx, fixture)
	if outboxRedacted != 1 || deadRedacted != 1 {
		t.Fatalf("safe redaction counts = outbox %d dead-letter %d",
			outboxRedacted, deadRedacted)
	}
	assertRetentionPayloads(
		t, ctx, fixture, rootID, normalRootID,
		uuid.MustParse(fixture.envelopes[2].EventID), normalEventID,
		true,
	)
	fixture.deliver(t, ctx, 1)
	if !fixture.gapReady(t, ctx, jobID) {
		t.Fatal("expired but protected live gap did not become ready")
	}
	claim := fixture.claimGap(t, ctx, jobID)
	assertRetiredGapRedactionEntrypoints(t, ctx, fixture, jobID, "running")
	outboxRedacted, deadRedacted = runGapSafeRedaction(t, ctx, fixture)
	if outboxRedacted != 0 || deadRedacted != 0 {
		t.Fatalf("running gap redaction counts = outbox %d dead-letter %d",
			outboxRedacted, deadRedacted)
	}
	assertRetentionPayloads(
		t, ctx, fixture, rootID, normalRootID,
		uuid.MustParse(fixture.envelopes[2].EventID), normalEventID,
		true,
	)
	result, err := fixture.store.ReconcileEventGap(ctx, claim)
	if err != nil || !result.Requeued() {
		t.Fatalf("reconcile protected expired gap = %#v, %v", result, err)
	}
	republishGap(t, ctx, fixture, 2)
	ageGapRetentionRows(
		t, ctx, fixture, rootID,
		uuid.MustParse(fixture.envelopes[2].EventID), uuid.Nil,
	)
	outboxRedacted, deadRedacted = runGapSafeRedaction(t, ctx, fixture)
	if outboxRedacted != 1 || deadRedacted != 1 {
		t.Fatalf("terminal redaction counts = outbox %d dead-letter %d",
			outboxRedacted, deadRedacted)
	}
	assertRetentionPayloads(
		t, ctx, fixture, rootID, normalRootID,
		uuid.MustParse(fixture.envelopes[2].EventID), normalEventID,
		false,
	)
}

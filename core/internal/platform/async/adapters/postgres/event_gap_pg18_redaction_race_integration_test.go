//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
)

func TestEventGapParkSerializesWithOutboxRedactionOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
			defer cancel()
			fixture := newEventGapPGFixture(t, ctx, origin)
			assertEventGapParkWinsRedactionRace(t, ctx, &fixture)
		})
	}
}

func assertEventGapParkWinsRedactionRace(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
) {
	t.Helper()
	fixture.seedDelivered(t, ctx, 1)
	fixture.seedDeliveredAt(t, ctx, 2, time.Now().UTC().Add(-32*24*time.Hour))
	eventID := fixture.envelopes[2].EventID
	ageOutboxForGapRedactionRace(t, ctx, fixture, eventID)
	releaseBarrier := installEventGapParkBarrier(t, ctx, fixture)
	defer releaseBarrier()

	delivery := make(chan eventGapRaceDelivery, 1)
	go func() {
		outcome, err := fixture.handler.Handle(
			ctx, fixture.envelopes[2], brokerrabbit.DeliveryMetadata{},
		)
		delivery <- eventGapRaceDelivery{outcome: outcome, err: err}
	}()
	waitForGapParkBarrier(t, ctx, fixture)

	assertGapRetentionRuntimeIdentity(t, ctx, fixture)
	redaction := make(chan eventGapRaceRedaction, 1)
	go runOutboxRedactionForGapRace(ctx, fixture, redaction)
	waitForGapRedactionLock(t, ctx, fixture)
	assertGapRaceHasNoPartialCommit(t, ctx, fixture, eventID)
	releaseBarrier()

	delivered := <-delivery
	if delivered.err != nil || delivered.outcome != brokerrabbit.HandleCommitted {
		t.Fatalf("gap park race delivery = %s, %v", delivered.outcome, delivered.err)
	}
	redacted := <-redaction
	if redacted.err != nil || redacted.count != 0 {
		t.Fatalf("gap race redaction = %d, %v", redacted.count, redacted.err)
	}
	assertGapRaceEvidence(t, ctx, fixture, eventID)
}

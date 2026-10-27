//go:build integration

package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type retentionCycleWorkIDs struct {
	redactJob    uuid.UUID
	purgeJob     uuid.UUID
	staleJob     uuid.UUID
	purgeInbox   uuid.UUID
	purgeReceipt uuid.UUID
	normalEvent  uuid.UUID
	normalRoot   uuid.UUID
	purgeRoot    uuid.UUID
	staleRoot    uuid.UUID
	liveEvent    uuid.UUID
	liveRoot     uuid.UUID
}

func TestAsyncRetentionCycleWorkOnPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
			defer cancel()
			fixture := newEventGapPGFixture(t, ctx, origin)
			ids := seedRetentionCycleWork(t, ctx, &fixture)
			acquired, counts := readRetentionCycle(t, ctx, fixture.runtime)
			if !acquired {
				t.Fatal("retention cycle did not acquire its transaction lock")
			}
			want := [7]int32{1, 1, 1, 1, 1, 1, 1}
			if counts != want {
				t.Fatalf("retention cycle counts = %v, want %v", counts, want)
			}
			assertRetentionCycleWork(t, ctx, fixture.owner, ids)
			assertRetentionCycleRollback(t, ctx, &fixture, ids.purgeInbox)
		})
	}
}

func seedRetentionCycleWork(
	t *testing.T, ctx context.Context, fixture *eventGapPGFixture,
) retentionCycleWorkIDs {
	t.Helper()
	fixture.seedDelivered(t, ctx, 1)
	fixture.seedDeliveredAt(t, ctx, 2, time.Now().UTC().Add(-32*24*time.Hour))
	fixture.deliver(t, ctx, 2)
	liveRoot, _ := fixture.gapIDs(t, ctx, 2)
	fixture.deliver(t, ctx, 1)
	normalEvent := seedExpiredNormalOutbox(t, ctx, fixture)
	normalRoot := insertExpiredNormalDeadLetter(t, ctx, fixture, liveRoot, normalEvent)
	liveEvent := uuid.MustParse(fixture.envelopes[2].EventID)
	ageGapRetentionRows(t, ctx, fixture, liveRoot, liveEvent, normalEvent)
	ids := newRetentionCycleWorkIDs(normalEvent, normalRoot, liveEvent, liveRoot)
	ids.purgeInbox = uuid.MustParse(fixture.envelopes[1].EventID)
	seedRetentionCycleRows(t, ctx, fixture.owner, ids,
		ids.purgeInbox)
	return ids
}

func newRetentionCycleWorkIDs(
	normalEvent, normalRoot, liveEvent, liveRoot uuid.UUID,
) retentionCycleWorkIDs {
	return retentionCycleWorkIDs{
		redactJob: uuid.New(), purgeJob: uuid.New(), staleJob: uuid.New(),
		purgeReceipt: uuid.New(), normalEvent: normalEvent, normalRoot: normalRoot,
		purgeRoot: uuid.New(), staleRoot: uuid.New(), liveEvent: liveEvent,
		liveRoot: liveRoot,
	}
}

func readRetentionCycle(
	t *testing.T, ctx context.Context, querier retentionCycleQuerier,
) (bool, [7]int32) {
	t.Helper()
	acquired, counts, err := queryRetentionCycle(ctx, querier)
	if err != nil {
		t.Fatalf("run retention cycle: %v", err)
	}
	return acquired, counts
}

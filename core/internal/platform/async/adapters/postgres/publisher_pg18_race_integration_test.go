//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type publisherRaceResult struct {
	claims []PublishClaim
	err    error
}

func assertPublisherClaimRace(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
) {
	t.Helper()
	tenantID := uuid.MustParse(asyncTenantA)
	eventID := uuid.New()
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: eventID, tenantID: tenantID,
		aggregateID: uuid.New(), sequence: 1,
	})
	authority := publisherAuthorityFor(t, ctx, fixture, tenantID)
	start := make(chan struct{})
	results := make(chan publisherRaceResult, 2)
	for range 2 {
		go func(publisherID uuid.UUID) {
			<-start
			claims, err := fixture.store.ClaimOutbox(
				ctx, authority, publisherID, 1, fixture.catalog,
			)
			results <- publisherRaceResult{claims: claims, err: err}
		}(uuid.New())
	}
	close(start)
	var winner PublishClaim
	total := 0
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatalf("publisher race: %v", result.err)
		}
		total += len(result.claims)
		if len(result.claims) == 1 {
			winner = result.claims[0]
		}
	}
	if total != 1 || winner.eventID != eventID {
		t.Fatalf("publisher race claimed %d rows, winner = %#v", total, winner)
	}
	confirmPublisherClaim(t, ctx, fixture, winner, "f")
}

package offering

import (
	"context"
	"testing"
)

func TestRetireCarriesCASAndReason(t *testing.T) {
	actor, offeringID := v7(t), v7(t)
	store := storeFixture{retire: func(value RetireMutation) (Snapshot, error) {
		if value.OfferingID != offeringID || value.ExpectedRevision != 6 ||
			value.Reason != "catalog replacement" || value.RequestSHA256 == [32]byte{} {
			t.Fatalf("mutation = %#v", value)
		}
		return Snapshot{OfferingID: offeringID, Revision: 7, Lifecycle: "retired"}, nil
	}}
	result, err := commandFixture(t, store).Retire(context.Background(), RetireRequest{
		ActorID: actor, IdempotencyKey: "offering-retire-1", OfferingID: offeringID,
		ExpectedRevision: 6, Reason: "catalog replacement",
	})
	if err != nil || result.Lifecycle != "retired" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestRetireRejectsBlankReason(t *testing.T) {
	commands := commandFixture(t, storeFixture{retire: func(RetireMutation) (Snapshot, error) {
		t.Fatal("store called")
		return Snapshot{}, nil
	}})
	_, err := commands.Retire(context.Background(), RetireRequest{
		ActorID: v7(t), IdempotencyKey: "key", OfferingID: v7(t), ExpectedRevision: 1,
	})
	if err != ErrRequest {
		t.Fatalf("expected ErrRequest, got %v", err)
	}
}

func testContext() context.Context { return context.Background() }

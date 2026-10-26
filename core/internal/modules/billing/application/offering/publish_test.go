package offering

import "testing"

func TestPublishCarriesExpectedRevision(t *testing.T) {
	actor, offeringID := v7(t), v7(t)
	store := storeFixture{createVersion: func(value CreateVersionMutation) (Snapshot, error) {
		if value.OfferingID != offeringID || value.ExpectedRevision != 4 ||
			value.RequestSHA256 == [32]byte{} || len(value.PriceIDs) != 1 {
			t.Fatalf("mutation = %#v", value)
		}
		return Snapshot{OfferingID: offeringID, VersionID: value.VersionID,
			Revision: 5, Version: 3, Lifecycle: "active"}, nil
	}}
	commands := commandFixture(t, store)
	result, err := commands.Publish(testContext(), CreateVersionRequest{
		ActorID: actor, IdempotencyKey: "offering-version-1", OfferingID: offeringID,
		ExpectedRevision: 4, Version: versionFixture(t),
	})
	if err != nil || result.Revision != 5 || result.Version != 3 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestPublishRejectsInvalidExpectedRevision(t *testing.T) {
	commands := commandFixture(t, storeFixture{createVersion: func(
		CreateVersionMutation,
	) (Snapshot, error) {
		t.Fatal("store called")
		return Snapshot{}, nil
	}})
	_, err := commands.Publish(testContext(), CreateVersionRequest{
		ActorID: v7(t), IdempotencyKey: "key", OfferingID: v7(t),
		Version: versionFixture(t),
	})
	if err != ErrRequest {
		t.Fatalf("expected ErrRequest, got %v", err)
	}
}

package offering

import (
	"testing"

	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

func TestCreateBuildsAtomicMutation(t *testing.T) {
	actor := v7(t)
	store := storeFixture{create: func(value CreateMutation) (Snapshot, error) {
		if !validV7(value.OperationID) || !validV7(value.OfferingID) ||
			!validV7(value.VersionID) || len(value.PriceIDs) != 1 ||
			value.ActorID != actor || value.RequestSHA256 == [32]byte{} ||
			value.Key != "pro" || value.IdempotencyKey != "offering-create-1" {
			t.Fatalf("mutation = %#v", value)
		}
		return Snapshot{OfferingID: value.OfferingID, VersionID: value.VersionID,
			Revision: 1, Version: 1, Lifecycle: "active"}, nil
	}}
	commands := commandFixture(t, store)
	key, _ := offeringdomain.NewKey("pro")
	result, err := commands.Create(testContext(), CreateRequest{ActorID: actor,
		IdempotencyKey: "offering-create-1", Key: key, Version: versionFixture(t)})
	if err != nil || result.Revision != 1 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestCreateRejectsUnvalidatedRequest(t *testing.T) {
	commands := commandFixture(t, storeFixture{create: func(CreateMutation) (Snapshot, error) {
		t.Fatal("store called")
		return Snapshot{}, nil
	}})
	if _, err := commands.Create(testContext(), CreateRequest{}); err != ErrRequest {
		t.Fatalf("expected ErrRequest, got %v", err)
	}
}

package discountcatalog

import (
	"context"
	"testing"
)

func TestCreateBuildsImmutableDiscountMutation(t *testing.T) {
	actor := v7(t)
	store := storeFixture{create: func(value CreateMutation) (Snapshot, error) {
		if !validV7(value.OperationID) || !validV7(value.DiscountID) ||
			!validV7(value.VersionID) || value.ActorID != actor ||
			value.IdempotencyKey != "discount-create-1" || value.RequestSHA256 == [32]byte{} ||
			value.Presentation.Name() != "Save ten" ||
			value.Presentation.Description() != "Loyal customers" || !value.Available {
			t.Fatalf("mutation = %#v", value)
		}
		return Snapshot{DiscountID: value.DiscountID, VersionID: value.VersionID,
			Revision: 1, Version: 1, Lifecycle: "active"}, nil
	}}
	result, err := commandsFixture(t, store).Create(context.Background(), CreateRequest{
		ActorID: actor, IdempotencyKey: "discount-create-1", Policy: policyFixture(t),
		Presentation: presentationFixture(t), Available: true,
	})
	if err != nil || result.Version != 1 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

func TestCreateRejectsZeroPolicy(t *testing.T) {
	commands := commandsFixture(t, storeFixture{create: func(CreateMutation) (Snapshot, error) {
		t.Fatal("store called")
		return Snapshot{}, nil
	}})
	if _, err := commands.Create(context.Background(), CreateRequest{}); err != ErrRequest {
		t.Fatalf("expected ErrRequest, got %v", err)
	}
}

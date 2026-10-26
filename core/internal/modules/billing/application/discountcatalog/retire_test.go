package discountcatalog

import (
	"context"
	"testing"
)

func TestRetireUsesExpectedRevision(t *testing.T) {
	actor, discountID := v7(t), v7(t)
	store := storeFixture{retire: func(value RetireMutation) (Snapshot, error) {
		if value.DiscountID != discountID || value.ExpectedRevision != 3 ||
			value.Reason != "promotion ended" {
			t.Fatalf("mutation = %#v", value)
		}
		return Snapshot{DiscountID: discountID, Revision: 4, Lifecycle: "revoked"}, nil
	}}
	result, err := commandsFixture(t, store).Retire(context.Background(), RetireRequest{
		ActorID: actor, IdempotencyKey: "discount-retire-1", DiscountID: discountID,
		ExpectedRevision: 3, Reason: "promotion ended",
	})
	if err != nil || result.Lifecycle != "revoked" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

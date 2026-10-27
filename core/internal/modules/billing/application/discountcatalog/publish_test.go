package discountcatalog

import (
	"context"
	"testing"
)

func TestPublishUsesExactCAS(t *testing.T) {
	actor, discountID := v7(t), v7(t)
	available := false
	store := storeFixture{createVersion: func(value CreateVersionMutation) (Snapshot, error) {
		if value.DiscountID != discountID || value.ExpectedRevision != 2 ||
			value.RequestSHA256 == [32]byte{} || value.Available == nil || *value.Available ||
			value.Presentation.Name() != "Save ten" ||
			value.Presentation.Description() != "Loyal customers" {
			t.Fatalf("mutation = %#v", value)
		}
		return Snapshot{DiscountID: discountID, VersionID: value.VersionID,
			Revision: 3, Version: 2, Lifecycle: "active"}, nil
	}}
	result, err := commandsFixture(t, store).Publish(
		context.Background(), CreateVersionRequest{ActorID: actor,
			IdempotencyKey: "discount-version-1", DiscountID: discountID,
			ExpectedRevision: 2, Policy: policyFixture(t),
			Presentation: presentationFixture(t), Available: &available},
	)
	if err != nil || result.Revision != 3 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

package discount

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
)

type testStore struct {
	snapshot EligibilitySnapshot
	existing Redemption
	found    bool
	created  []Redemption
}

func (store *testStore) Redeem(
	_ context.Context,
	_ RedeemExecution,
	planner RedeemPlanner,
) (RedeemResult, error) {
	var existing *Redemption
	if store.found {
		existing = &store.existing
	}
	plan, err := planner(store.snapshot, existing)
	if err != nil {
		return RedeemResult{}, err
	}
	if plan.Create != nil {
		store.created = append(store.created, *plan.Create)
	}
	return plan.Result, nil
}

func TestRedeemPinsExactPolicyAndVersions(t *testing.T) {
	commands, store, request, redemptionID, _ := redeemFixture(t)
	result, err := commands.Redeem(context.Background(), request)
	if err != nil || result.Replay || result.Redemption.ID != redemptionID ||
		result.Redemption.PolicySHA256 != store.snapshot.PolicySHA256 ||
		len(store.created) != 1 {
		t.Fatalf("result=%+v created=%+v err=%v", result, store.created, err)
	}
}

func TestRedeemRejectsExhaustedLimit(t *testing.T) {
	commands, store, request, _, _ := redeemFixture(t)
	limit := int64(2)
	store.snapshot.GlobalUseLimit = &limit
	store.snapshot.GlobalUsed = 2
	if _, err := commands.Redeem(context.Background(), request); err != ErrIneligible {
		t.Fatalf("expected ErrIneligible, got %v", err)
	}
	if len(store.created) != 0 {
		t.Fatal("exhausted redemption was written")
	}
}

func TestRedeemExactOperationReplays(t *testing.T) {
	commands, store, request, redemptionID, now := redeemFixture(t)
	store.found = true
	store.existing = Redemption{
		ID: redemptionID, DiscountVersionID: store.snapshot.DiscountVersionID,
		SubscriptionID:    request.SubscriptionID,
		OfferingVersionID: request.OfferingVersionID,
		PriceVersionID:    request.PriceVersionID,
		SourceOperationID: request.SourceOperationID,
		PolicySHA256:      store.snapshot.PolicySHA256, RedeemedAt: now,
	}
	result, err := commands.Redeem(context.Background(), request)
	if err != nil || !result.Replay || len(store.created) != 0 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestRedeemRejectsOperationReuse(t *testing.T) {
	commands, store, request, redemptionID, now := redeemFixture(t)
	store.found = true
	store.existing = Redemption{
		ID: redemptionID, DiscountVersionID: store.snapshot.DiscountVersionID,
		SubscriptionID:    request.SubscriptionID,
		OfferingVersionID: request.OfferingVersionID,
		PriceVersionID:    request.PriceVersionID,
		SourceOperationID: request.SourceOperationID,
		PolicySHA256:      sha256.Sum256([]byte("different")), RedeemedAt: now,
	}
	if _, err := commands.Redeem(context.Background(), request); !errors.Is(
		err, ErrOperationReused,
	) {
		t.Fatalf("expected ErrOperationReused, got %v", err)
	}
}

func redeemFixture(
	t *testing.T,
) (*Commands, *testStore, RedeemRequest, uuid.UUID, time.Time) {
	t.Helper()
	now := time.Date(2026, time.July, 17, 16, 0, 0, 0, time.UTC)
	code, _ := discountdomain.NormalizeCode("launch_2026")
	offeringID, priceID := mustV7(t), mustV7(t)
	store := &testStore{snapshot: EligibilitySnapshot{
		DiscountVersionID: mustV7(t), Code: code,
		OfferingVersionID: offeringID, PriceVersionID: priceID,
		RequiredCurrency: "USD", ValidFrom: now.Add(-time.Hour),
		ValidUntil: now.Add(time.Hour), OfferingAllowed: true,
		PolicySHA256: sha256.Sum256([]byte("policy")),
	}}
	redemptionID := mustV7(t)
	commands, err := NewCommands(
		store, func() (uuid.UUID, error) { return redemptionID, nil },
		func() time.Time { return now },
	)
	if err != nil {
		t.Fatal(err)
	}
	request := RedeemRequest{
		Code: code, SubscriptionID: mustV7(t),
		OfferingVersionID: offeringID, PriceVersionID: priceID,
		Currency: "USD", SourceOperationID: mustV7(t),
	}
	return commands, store, request, redemptionID, now
}

func mustV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

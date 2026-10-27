package entitlement

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

type testStore struct{ snapshot Snapshot }

func (store testStore) LoadEffectiveSnapshot(context.Context, time.Time) (Snapshot, error) {
	return store.snapshot, nil
}

func TestResolveAppliesBasePurchasedAndOverridePrecedence(t *testing.T) {
	now := time.Date(2026, time.July, 17, 15, 0, 0, 0, time.UTC)
	key, definition := definitionFixture(t)
	base, _ := domain.IntegerValue(10)
	purchased, _ := domain.IntegerValue(20)
	override, _ := domain.IntegerValue(30)
	window, _ := billingperiod.New(now.Add(-time.Hour), now.Add(time.Hour))
	snapshot := Snapshot{
		Definitions:      map[domain.Key]domain.Definition{key: definition},
		BaseAssignmentID: mustV7(t), BaseOfferingVersionID: mustV7(t),
		Base: map[domain.Key]domain.Value{key: base},
		Purchased: &PurchasedLayer{
			SubscriptionID: mustV7(t), OfferingVersionID: mustV7(t),
			State: subscription.Active, GrantWindow: &window,
			Values:    map[domain.Key]domain.Value{key: purchased},
			Overrides: map[domain.Key]domain.Value{key: override},
		},
	}
	queries, _ := NewQueries(testStore{snapshot}, func() time.Time { return now })
	result, err := queries.Resolve(context.Background())
	value, ok := result.Values()[key].Integer()
	if err != nil || !ok || value != 30 || result.SubscriptionID() == nil {
		t.Fatalf("value=%d ok=%v result=%+v err=%v", value, ok, result, err)
	}
}

func TestResolveDropsPurchasedGrantAtBoundary(t *testing.T) {
	now := time.Date(2026, time.July, 17, 15, 0, 0, 0, time.UTC)
	key, definition := definitionFixture(t)
	base, _ := domain.IntegerValue(10)
	purchased, _ := domain.IntegerValue(20)
	window, _ := billingperiod.New(now.Add(-time.Hour), now)
	snapshot := Snapshot{
		Definitions:      map[domain.Key]domain.Definition{key: definition},
		BaseAssignmentID: mustV7(t), BaseOfferingVersionID: mustV7(t),
		Base: map[domain.Key]domain.Value{key: base},
		Purchased: &PurchasedLayer{
			SubscriptionID: mustV7(t), OfferingVersionID: mustV7(t),
			State: subscription.Active, GrantWindow: &window,
			Values:    map[domain.Key]domain.Value{key: purchased},
			Overrides: map[domain.Key]domain.Value{},
		},
	}
	queries, _ := NewQueries(testStore{snapshot}, func() time.Time { return now })
	result, err := queries.Resolve(context.Background())
	value, _ := result.Values()[key].Integer()
	if err != nil || value != 10 || result.SubscriptionID() != nil {
		t.Fatalf("value=%d result=%+v err=%v", value, result, err)
	}
}

func definitionFixture(t *testing.T) (domain.Key, domain.Definition) {
	t.Helper()
	key, _ := domain.NewKey("iam.active_users")
	definition, err := domain.NewDefinition(
		key, domain.Integer, domain.Counter, domain.BillingPeriod, domain.Hard, true)
	if err != nil {
		t.Fatal(err)
	}
	return key, definition
}

func mustV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

package subscriptionapp

import (
	"context"
	"errors"
	"testing"
	"time"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type rejectingCurrentStore struct{ called bool }

func (store *rejectingCurrentStore) GetCurrent(context.Context) (CurrentSubscription, error) {
	store.called = true
	return CurrentSubscription{}, ErrSubscriptionQueryStorage
}

func TestGetCurrentRejectsUnboundScopeAndValidatesImmutableSnapshot(t *testing.T) {
	store := &rejectingCurrentStore{}
	queries, err := NewQueries(store)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = queries.GetCurrent(
		context.Background(), platformtx.TenantScope{},
	); !errors.Is(err, ErrQueryScope) || store.called {
		t.Fatalf("unbound query = %v, called = %t", err, store.called)
	}
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	period, err := billingperiod.New(now.Add(-time.Hour), now.AddDate(0, 1, 0))
	if err != nil {
		t.Fatal(err)
	}
	offeringID, priceID := uuid.New(), uuid.New()
	offering, err := NewCurrentOffering(CurrentOfferingInput{
		ID: uuid.New(), VersionID: offeringID, Key: "team", State: "retired",
		Name: "Historical Team", Revision: 2, Version: 1, BillingInterval: "monthly",
		ExpectedPriceCount: 1, EffectiveStart: now.Add(-24 * time.Hour),
		CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour),
		Prices: []CurrentOfferingPrice{{ID: priceID, Cadence: "monthly",
			MinorUnits: 9900, Currency: "USD", CurrencyRegistryVersion: 1}},
	})
	if err != nil {
		t.Fatal(err)
	}
	identifier := uuid.New()
	value, err := NewCurrentSubscription(CurrentSubscriptionInput{
		ID: identifier, LegacyID: identifier, LegacyPlanID: offering.ID(),
		TenantID: uuid.New(), ProviderAccountID: uuid.New(),
		OfferingVersionID: offeringID, PriceVersionID: priceID, Provider: "manual",
		State: subscriptiondomain.Active, Period: &period,
		EffectiveMinor: 9900, EffectiveCurrency: "USD", EffectiveRegistry: 1,
		Revision: 3, CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now,
		Offering: offering,
	})
	if err != nil || value.Provider() != "manual" || value.Revision() != 3 ||
		value.Offering().State() != "retired" || value.Offering().Public() {
		t.Fatalf("current subscription = %#v, %v", value, err)
	}
	copyPeriod, available := value.Period()
	if !available || !copyPeriod.End().Equal(period.End()) {
		t.Fatalf("period = %#v, available = %t", copyPeriod, available)
	}
}

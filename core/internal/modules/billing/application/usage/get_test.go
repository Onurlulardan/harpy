package usage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type rejectingUsageStore struct{ called bool }

func (store *rejectingUsageStore) Get(context.Context) (UsageSnapshot, error) {
	store.called = true
	return UsageSnapshot{}, ErrQueryStorage
}

func TestGetRejectsUnboundScopeAndBuildsTypedCounterSnapshot(t *testing.T) {
	store := &rejectingUsageStore{}
	queries, err := NewQueries(store)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = queries.Get(
		context.Background(), platformtx.TenantScope{},
	); !errors.Is(err, ErrQueryScope) || store.called {
		t.Fatalf("unbound query = %v, called = %t", err, store.called)
	}
	key, err := entitlement.NewKey("forms.submissions")
	if err != nil {
		t.Fatal(err)
	}
	definition, err := entitlement.NewDefinition(
		key, entitlement.Integer, entitlement.Counter,
		entitlement.CalendarMonthUTC, entitlement.Hard, true,
	)
	if err != nil {
		t.Fatal(err)
	}
	limit, err := entitlement.IntegerValue(10)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	period, err := billingperiod.New(now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	meter, err := NewMeterSnapshot(MeterSnapshotInput{
		Definition: definition, Limit: limit, PeriodID: uuid.Must(uuid.NewV7()),
		Period: period, Committed: 4, Reserved: 2,
		PeriodRevision: 1, TotalRevision: 2, UpdatedAt: now,
	})
	if err != nil || meter.Current() != 6 {
		t.Fatalf("meter current = %d, error = %v", meter.Current(), err)
	}
	snapshot, err := NewUsageSnapshot(uuid.New(), []MeterSnapshot{meter})
	if err != nil || len(snapshot.Meters()) != 1 {
		t.Fatalf("usage snapshot meters = %d, error = %v", len(snapshot.Meters()), err)
	}
}

func TestGaugeSnapshotCurrentIncludesPendingReservations(t *testing.T) {
	key, _ := entitlement.NewKey("documents.storage_bytes")
	definition, err := entitlement.NewDefinition(key, entitlement.Integer,
		entitlement.Gauge, entitlement.Never, entitlement.Hard, true)
	if err != nil {
		t.Fatal(err)
	}
	limit, _ := entitlement.IntegerValue(100)
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	period, _ := billingperiod.New(now.Add(-time.Hour), now.Add(time.Hour))
	meter, err := NewMeterSnapshot(MeterSnapshotInput{
		Definition: definition, Limit: limit, PeriodID: uuid.Must(uuid.NewV7()),
		Period: period, Reserved: 3, Gauge: 4, PeriodRevision: 1,
		TotalRevision: 2, UpdatedAt: now,
	})
	if err != nil || meter.Current() != 7 {
		t.Fatalf("gauge current=%d error=%v", meter.Current(), err)
	}
}

package subscriptionapp

import (
	"testing"
	"time"

	"github.com/google/uuid"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

var fixtureNow = time.Date(2026, time.July, 17, 10, 0, 0, 0, time.UTC)

func fixtureCommands(t *testing.T, store *fakeStore, now time.Time) *Commands {
	t.Helper()
	if store.receipts == nil {
		store.receipts = map[uuid.UUID]OperationReceipt{}
	}
	commands, err := NewCommands(store, uuid.NewV7, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	return commands
}

func fixtureSnapshot(t *testing.T, state subscriptiondomain.State) SubscriptionSnapshot {
	t.Helper()
	start := fixtureNow.Add(-time.Hour)
	end := fixtureNow.Add(23 * time.Hour)
	periodValue := fixturePeriod(t, start, end)
	snapshot := SubscriptionSnapshot{
		ID: uuid.New(), TenantID: uuid.New(), ProviderAccountID: uuid.New(),
		OfferingVersionID: uuid.New(), PriceVersionID: uuid.New(), Currency: "TRY",
		State: state, Revision: 4, UpdatedAt: start, Period: &periodValue,
	}
	switch state {
	case subscriptiondomain.Pending:
		snapshot.Period = nil
	case subscriptiondomain.Trialing:
		trialEnd := end
		snapshot.TrialEndsAt = &trialEnd
	case subscriptiondomain.PastDue:
		graceEnd := end
		snapshot.GraceEndsAt = &graceEnd
	case subscriptiondomain.Cancelled, subscriptiondomain.Expired,
		subscriptiondomain.Paused:
		snapshot.Period = nil
	}
	if !snapshot.valid() {
		t.Fatal("fixture subscription snapshot is invalid")
	}
	return snapshot
}

func fixturePeriod(t *testing.T, start, end time.Time) billingperiod.Period {
	t.Helper()
	value, err := billingperiod.New(start, end)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func fixtureTarget(snapshot SubscriptionSnapshot) TargetPrice {
	return TargetPrice{
		OfferingVersionID: uuid.New(), PriceVersionID: uuid.New(),
		Currency: snapshot.Currency, TrialDays: 14,
	}
}

func fixtureV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func fixtureJob(
	t *testing.T,
	snapshot SubscriptionSnapshot,
	kind BoundaryKind,
	runAt time.Time,
) JobIntent {
	t.Helper()
	return JobIntent{
		ID: fixtureV7(t), EffectID: fixtureV7(t), SubscriptionID: snapshot.ID,
		Boundary: kind, RunAt: runAt, ExpectedRevision: snapshot.Revision,
		PayloadRevision: 1,
	}
}

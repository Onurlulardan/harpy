package usage

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
)

func gaugeFixture(
	t *testing.T,
	policy entitlement.Enforcement,
	current int64,
	limitValue int64,
) (*Commands, *testStore, AdjustGaugeRequest, []uuid.UUID) {
	t.Helper()
	now := time.Date(2026, time.July, 17, 14, 0, 0, 0, time.UTC)
	key, _ := entitlement.NewKey("documents.bytes")
	definition, err := entitlement.NewDefinition(
		key, entitlement.Integer, entitlement.Gauge,
		entitlement.Never, policy, true,
	)
	if err != nil {
		t.Fatal(err)
	}
	limit, _ := entitlement.IntegerValue(limitValue)
	periodValue, err := billingperiod.New(now.Add(-time.Hour), now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	periodID := mustV7(t)
	store := &testStore{gaugeSnapshot: GaugeSnapshot{
		Definition: definition, Limit: limit, PeriodID: periodID,
		Period: periodValue, Current: current, Revision: 11,
	}}
	identifiers := []uuid.UUID{mustV7(t), mustV7(t), mustV7(t)}
	index := 0
	commands, err := NewCommands(
		store,
		func() (uuid.UUID, error) {
			identifier := identifiers[index]
			index++
			return identifier, nil
		},
		func() time.Time { return now },
	)
	if err != nil {
		t.Fatal(err)
	}
	request := AdjustGaugeRequest{
		EntitlementKey: key, PeriodID: periodID, OwnerFactID: mustV7(t), Delta: 2,
	}
	return commands, store, request, identifiers
}

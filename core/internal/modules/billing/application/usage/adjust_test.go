package usage

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

func TestAdjustGaugeWritesDecisionThenOwnerFactEntry(t *testing.T) {
	commands, store, request, identifiers := gaugeFixture(t, entitlement.Hard, 3, 10)
	result, err := commands.AdjustGauge(context.Background(), request)
	if err != nil || !result.Allowed || result.Before != 3 || result.After != 5 ||
		result.Entry.ID != identifiers[1] || result.Entry.DecisionID != identifiers[0] {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
	if !reflect.DeepEqual(store.writeOrder, []string{"decision", "gauge"}) ||
		len(store.gaugeEntries) != 1 {
		t.Fatalf("writes=%v entries=%d", store.writeOrder, len(store.gaugeEntries))
	}
}

func TestAdjustGaugeHardDenialDoesNotAppendEntry(t *testing.T) {
	commands, store, request, _ := gaugeFixture(t, entitlement.Hard, 9, 10)
	result, err := commands.AdjustGauge(context.Background(), request)
	if err != nil || result.Allowed || result.After != 11 ||
		len(store.decisions) != 1 || len(store.gaugeEntries) != 0 {
		t.Fatalf("result=%+v decisions=%d entries=%d err=%v",
			result, len(store.decisions), len(store.gaugeEntries), err)
	}
}

func TestAdjustGaugeSoftLimitEmitsCrossingEvent(t *testing.T) {
	commands, store, request, identifiers := gaugeFixture(t, entitlement.Soft, 9, 10)
	result, err := commands.AdjustGauge(context.Background(), request)
	if err != nil || !result.Allowed || !result.SoftLimitExceeded ||
		len(store.events) != 1 || store.events[0].ID != identifiers[2] ||
		store.events[0].UsageRevision != store.gaugeSnapshot.Revision+1 {
		t.Fatalf("result=%+v events=%+v err=%v", result, store.events, err)
	}
}

func TestAdjustGaugeExactOwnerFactReplays(t *testing.T) {
	commands, store, request, identifiers := gaugeFixture(t, entitlement.Hard, 5, 10)
	now := commands.clock()
	store.gaugeFound = true
	store.gaugeReplay = GaugeReplay{
		Entry: GaugeEntry{
			ID: identifiers[1], DecisionID: identifiers[0],
			EntitlementKey: request.EntitlementKey, PeriodID: request.PeriodID,
			OwnerFactID: request.OwnerFactID, Delta: request.Delta,
			Before: 3, After: 5, CreatedAt: now,
		},
		Decision: DecisionRecord{
			ID: identifiers[0], EntitlementKey: request.EntitlementKey,
			PeriodID: request.PeriodID, SourceOperationID: request.OwnerFactID,
			Current: 3, Requested: request.Delta, Limit: store.gaugeSnapshot.Limit,
			Allowed: true, DecidedAt: now,
		},
	}
	result, err := commands.AdjustGauge(context.Background(), request)
	if err != nil || !result.Replay || result.After != 5 || len(store.writeOrder) != 0 {
		t.Fatalf("result=%+v writes=%v err=%v", result, store.writeOrder, err)
	}
}

func TestAdjustGaugeRejectsOwnerFactReuse(t *testing.T) {
	commands, store, request, identifiers := gaugeFixture(t, entitlement.Hard, 5, 10)
	store.gaugeFound = true
	store.gaugeReplay = GaugeReplay{
		Entry: GaugeEntry{ID: identifiers[1], OwnerFactID: request.OwnerFactID,
			Delta: request.Delta + 1},
	}
	if _, err := commands.AdjustGauge(context.Background(), request); !errors.Is(
		err, ErrOperationReused,
	) {
		t.Fatalf("expected ErrOperationReused, got %v", err)
	}
}

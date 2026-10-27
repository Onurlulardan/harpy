package usage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
)

type testStore struct {
	snapshot       ReservationSnapshot
	resolvedPeriod uuid.UUID
	resolveCalls   int
	reserveInput   ReserveExecution
	replay         ReservationReplay
	found          bool
	locked         Reservation
	lockedFound    bool
	created        []Reservation
	decisions      []DecisionRecord
	effectStarts   []EffectStart
	transitions    []ReservationTransition
	gaugeSnapshot  GaugeSnapshot
	gaugeReplay    GaugeReplay
	gaugeFound     bool
	gaugeEntries   []GaugeEntry
	events         []ThresholdEventIntent
	writeOrder     []string
}

func (store *testStore) ResolvePeriod(
	_ context.Context, _ entitlement.Key,
) (uuid.UUID, error) {
	store.resolveCalls++
	if store.resolvedPeriod != uuid.Nil {
		return store.resolvedPeriod, nil
	}
	return store.snapshot.PeriodID, nil
}

func (store *testStore) Reserve(
	_ context.Context, execution ReserveExecution, planner ReservePlanner,
) (ReserveResult, error) {
	store.reserveInput = execution
	var existing *ReservationReplay
	if store.found {
		existing = &store.replay
	}
	plan, err := planner(store.snapshot, existing)
	if err != nil {
		return ReserveResult{}, err
	}
	if plan.Decision != nil {
		store.decisions = append(store.decisions, *plan.Decision)
		store.writeOrder = append(store.writeOrder, "decision")
	}
	if plan.Reservation != nil {
		store.created = append(store.created, *plan.Reservation)
		store.writeOrder = append(store.writeOrder, "reservation")
	}
	if plan.Event != nil {
		store.events = append(store.events, *plan.Event)
	}
	return plan.Result, nil
}

func (store *testStore) Adjust(
	_ context.Context, _ AdjustExecution, planner AdjustPlanner,
) (AdjustGaugeResult, error) {
	var existing *GaugeReplay
	if store.gaugeFound {
		existing = &store.gaugeReplay
	}
	plan, err := planner(store.gaugeSnapshot, existing)
	if err != nil {
		return AdjustGaugeResult{}, err
	}
	if plan.Decision != nil {
		store.decisions = append(store.decisions, *plan.Decision)
		store.writeOrder = append(store.writeOrder, "decision")
	}
	if plan.Entry != nil {
		store.gaugeEntries = append(store.gaugeEntries, *plan.Entry)
		store.writeOrder = append(store.writeOrder, "gauge")
	}
	if plan.Event != nil {
		store.events = append(store.events, *plan.Event)
	}
	return plan.Result, nil
}

func (store *testStore) Start(
	_ context.Context, _ StartExecution, planner ReservationPlanner,
) (ReservationResult, error) {
	return store.applyReservationPlan(planner)
}

func (store *testStore) Complete(
	_ context.Context, _ TransitionExecution, planner ReservationPlanner,
) (ReservationResult, error) {
	return store.applyReservationPlan(planner)
}

func (store *testStore) Release(
	_ context.Context, _ TransitionExecution, planner ReservationPlanner,
) (ReservationResult, error) {
	return store.applyReservationPlan(planner)
}

func (store *testStore) Expire(
	_ context.Context, _ ExpireExecution, planner ReservationPlanner,
) (ReservationResult, error) {
	return store.applyReservationPlan(planner)
}

func (store *testStore) applyReservationPlan(
	planner ReservationPlanner,
) (ReservationResult, error) {
	plan, err := planner(store.locked, store.lockedFound)
	if err != nil {
		return ReservationResult{}, err
	}
	if plan.Effect != nil {
		store.effectStarts = append(store.effectStarts, *plan.Effect)
	}
	if plan.Transition != nil {
		store.transitions = append(store.transitions, *plan.Transition)
	}
	return plan.Result, nil
}

func reserveFixture(
	t *testing.T,
	enforcement entitlement.Enforcement,
	committed int64,
	reserved int64,
	limitValue int64,
) (*Commands, *testStore, ReserveRequest, []uuid.UUID) {
	t.Helper()
	now := time.Date(2026, time.July, 17, 10, 0, 0, 0, time.UTC)
	key, _ := entitlement.NewKey("iam.active_users")
	definition, err := entitlement.NewDefinition(
		key, entitlement.Integer, entitlement.Counter,
		entitlement.BillingPeriod, enforcement, true,
	)
	if err != nil {
		t.Fatal(err)
	}
	limit, _ := entitlement.IntegerValue(limitValue)
	periodValue, err := billingperiod.New(now, now.Add(24*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	periodID := mustV7(t)
	store := &testStore{snapshot: ReservationSnapshot{
		Definition: definition, Limit: limit, PeriodID: periodID,
		Period: periodValue, Committed: committed, Reserved: reserved, Revision: 7,
	}}
	identifiers := []uuid.UUID{mustV7(t), mustV7(t), mustV7(t)}
	index := 0
	commands, err := NewCommands(store, func() (uuid.UUID, error) {
		identifier := identifiers[index]
		index++
		return identifier, nil
	}, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	request := ReserveRequest{
		EntitlementKey: key, PeriodID: periodID, SourceOperationID: mustV7(t),
		Amount: 2, ExpiresAt: now.Add(time.Hour),
	}
	return commands, store, request, identifiers
}

func mustV7(t *testing.T) uuid.UUID {
	t.Helper()
	identifier, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return identifier
}

package usage

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func TestReserveWritesDecisionThenReservation(t *testing.T) {
	commands, store, request, identifiers := reserveFixture(
		t, entitlement.Hard, 3, 1, 10)
	result, err := commands.Reserve(context.Background(), request)
	if err != nil || !result.Allowed || result.Replay || result.Current != 4 ||
		result.Reservation.ID != identifiers[1] ||
		result.Reservation.DecisionID != identifiers[0] {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
	if !reflect.DeepEqual(store.writeOrder, []string{"decision", "reservation"}) ||
		len(store.decisions) != 1 || len(store.created) != 1 {
		t.Fatalf("writes=%v decisions=%d reservations=%d",
			store.writeOrder, len(store.decisions), len(store.created))
	}
	if store.created[0].State != usagedomain.Reserved ||
		store.created[0].SourceOperationID != request.SourceOperationID {
		t.Fatalf("unexpected reservation: %+v", store.created[0])
	}
}

func TestReserveHardDenialWritesOnlyDecision(t *testing.T) {
	commands, store, request, _ := reserveFixture(t, entitlement.Hard, 8, 1, 10)
	result, err := commands.Reserve(context.Background(), request)
	if err != nil || result.Allowed || result.Current != 9 || result.Requested != 2 {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
	if len(store.decisions) != 1 || len(store.created) != 0 ||
		!reflect.DeepEqual(store.writeOrder, []string{"decision"}) {
		t.Fatalf("unexpected writes: %v", store.writeOrder)
	}
}

func TestReserveSoftLimitEmitsCrossingEvent(t *testing.T) {
	commands, store, request, identifiers := reserveFixture(t, entitlement.Soft, 8, 1, 10)
	result, err := commands.Reserve(context.Background(), request)
	if err != nil || !result.Allowed || !result.SoftLimitExceeded ||
		len(store.events) != 1 || store.events[0].ID != identifiers[2] ||
		store.events[0].UsageRevision != store.snapshot.Revision+1 {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
}

func TestReserveAboveSoftLimitDoesNotRepeatThresholdEvent(t *testing.T) {
	commands, store, request, _ := reserveFixture(t, entitlement.Soft, 11, 0, 10)
	result, err := commands.Reserve(context.Background(), request)
	if err != nil || !result.SoftLimitExceeded || len(store.events) != 0 {
		t.Fatalf("result=%+v events=%+v err=%v", result, store.events, err)
	}
}

func TestReserveExactReplayDoesNotWrite(t *testing.T) {
	commands, store, request, identifiers := reserveFixture(t, entitlement.Hard, 3, 1, 10)
	store.found = true
	store.replay = ReservationReplay{
		Reservation: Reservation{
			ID: identifiers[1], DecisionID: identifiers[0],
			EntitlementKey: request.EntitlementKey, PeriodID: request.PeriodID,
			SourceOperationID: request.SourceOperationID, Amount: request.Amount,
			State: usagedomain.Committed, ExpiresAt: request.ExpiresAt,
			CommittedAmount: request.Amount, Revision: 2,
		},
		Decision: DecisionRecord{
			ID: identifiers[0], EntitlementKey: request.EntitlementKey,
			PeriodID: request.PeriodID, SourceOperationID: request.SourceOperationID,
			Current: 4, Requested: request.Amount, Limit: store.snapshot.Limit,
			Allowed: true,
		},
	}
	result, err := commands.Reserve(context.Background(), request)
	if err != nil || !result.Allowed || !result.Replay ||
		result.Reservation.State != usagedomain.Committed {
		t.Fatalf("result=%+v, err=%v", result, err)
	}
	if len(store.writeOrder) != 0 {
		t.Fatalf("replay wrote state: %v", store.writeOrder)
	}
}

func TestReserveRejectsOperationReuse(t *testing.T) {
	commands, store, request, identifiers := reserveFixture(t, entitlement.Hard, 3, 1, 10)
	store.found = true
	store.replay = ReservationReplay{
		Reservation: Reservation{
			ID: identifiers[1], DecisionID: identifiers[0],
			EntitlementKey: request.EntitlementKey, PeriodID: request.PeriodID,
			SourceOperationID: request.SourceOperationID, Amount: request.Amount + 1,
			State: usagedomain.Reserved, ExpiresAt: request.ExpiresAt,
		},
		Decision: DecisionRecord{ID: identifiers[0], Allowed: true},
	}
	if _, err := commands.Reserve(context.Background(), request); !errors.Is(
		err, ErrReservationReused,
	) {
		t.Fatalf("expected ErrReservationReused, got %v", err)
	}
}

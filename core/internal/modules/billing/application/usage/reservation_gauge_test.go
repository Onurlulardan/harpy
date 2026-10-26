package usage

import (
	"context"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	"github.com/google/uuid"
)

func TestReserveGaugeCountsCurrentAndPendingBytes(t *testing.T) {
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	key, _ := entitlement.NewKey("documents.storage_bytes")
	definition, err := entitlement.NewDefinition(key, entitlement.Integer,
		entitlement.Gauge, entitlement.Never, entitlement.Hard, true)
	if err != nil {
		t.Fatal(err)
	}
	limit, _ := entitlement.IntegerValue(100)
	period, _ := billingperiod.New(now.Add(-time.Hour), now.Add(time.Hour))
	periodID := mustV7(t)
	store := &testStore{snapshot: ReservationSnapshot{Definition: definition,
		Limit: limit, PeriodID: periodID, Period: period, Reserved: 10,
		Gauge: 70, Revision: 4}}
	commands, err := NewCommands(store, func() (uuid.UUID, error) {
		return mustV7(t), nil
	}, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	result, err := commands.Reserve(context.Background(), ReserveRequest{
		EntitlementKey: key, PeriodID: periodID, SourceOperationID: mustV7(t),
		Amount: 21, ExpiresAt: now.Add(time.Hour),
	})
	if err != nil || result.Allowed || result.Current != 80 || result.Requested != 21 {
		t.Fatalf("result=%+v error=%v", result, err)
	}
}

func TestReserveResolvesCurrentPeriodAndUsesSuppliedReservation(t *testing.T) {
	commands, store, request, _ := reserveFixture(t, entitlement.Hard, 1, 0, 10)
	reservationID := mustV7(t)
	request.PeriodID = uuid.Nil
	request.ResolveCurrentPeriod = true
	request.ReservationID = reservationID
	result, err := commands.Reserve(context.Background(), request)
	if err != nil || !result.Allowed || result.Reservation.ID != reservationID ||
		store.resolveCalls != 1 || store.reserveInput.Request.PeriodID != store.snapshot.PeriodID ||
		store.reserveInput.Request.ResolveCurrentPeriod {
		t.Fatalf("result=%+v calls=%d execution=%+v error=%v",
			result, store.resolveCalls, store.reserveInput, err)
	}
}

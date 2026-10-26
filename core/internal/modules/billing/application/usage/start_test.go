package usage

import (
	"context"
	"errors"
	"testing"
)

func TestStartEffectWritesOnceAndReplaysExactReceipt(t *testing.T) {
	commands, store, reservation, now := lifecycleFixture(t)
	receipt := mustV7(t)
	result, err := commands.StartEffect(context.Background(), StartEffectRequest{
		ReservationID: reservation.ID, ReceiptID: receipt,
	})
	if err != nil || result.Replay || len(store.effectStarts) != 1 ||
		result.Reservation.EffectStartedAt == nil ||
		!result.Reservation.EffectStartedAt.Equal(now) {
		t.Fatalf("result=%+v starts=%+v err=%v", result, store.effectStarts, err)
	}
	store.locked = result.Reservation
	replay, err := commands.StartEffect(context.Background(), StartEffectRequest{
		ReservationID: reservation.ID, ReceiptID: receipt,
	})
	if err != nil || !replay.Replay || len(store.effectStarts) != 1 {
		t.Fatalf("replay=%+v starts=%d err=%v", replay, len(store.effectStarts), err)
	}
}

func TestStartEffectRejectsReceiptReuse(t *testing.T) {
	commands, store, reservation, now := lifecycleFixture(t)
	storedReceipt := mustV7(t)
	store.locked.EffectStartedAt = instantPointer(now)
	store.locked.EffectReceiptID = identifierPointer(storedReceipt)
	_, err := commands.StartEffect(context.Background(), StartEffectRequest{
		ReservationID: reservation.ID, ReceiptID: mustV7(t),
	})
	if !errors.Is(err, ErrReservationReused) {
		t.Fatalf("expected ErrReservationReused, got %v", err)
	}
}

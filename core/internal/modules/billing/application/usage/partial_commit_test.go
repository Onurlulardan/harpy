package usage

import (
	"context"
	"errors"
	"testing"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func TestCompleteCommitsActualAmountAndReleasesRemainder(t *testing.T) {
	commands, store, reservation, _ := lifecycleFixture(t)
	reservation.Amount = 10
	store.locked = reservation
	receipt := mustV7(t)
	request := TransitionRequest{ReservationID: reservation.ID,
		ReceiptID: identifierPointer(receipt), CommittedAmount: 7}
	result, err := commands.Complete(context.Background(), request)
	if err != nil || result.Reservation.State != usagedomain.Committed ||
		result.Reservation.CommittedAmount != 7 || len(store.transitions) != 1 {
		t.Fatalf("result=%+v transitions=%+v error=%v", result, store.transitions, err)
	}
	change := store.transitions[0]
	if change.CommittedAmount != 7 || !validV7(change.RemainderEntryID) ||
		change.RemainderEntryID == change.LedgerEntryID {
		t.Fatalf("partial transition=%+v", change)
	}
	store.locked = result.Reservation
	replay, err := commands.Complete(context.Background(), request)
	if err != nil || !replay.Replay {
		t.Fatalf("replay=%+v error=%v", replay, err)
	}
	request.CommittedAmount = 8
	if _, err = commands.Complete(context.Background(), request); !errors.Is(
		err, ErrReservationReused,
	) {
		t.Fatalf("expected amount conflict, got %v", err)
	}
}

func TestCompleteRejectsAmountAboveReservation(t *testing.T) {
	commands, _, reservation, _ := lifecycleFixture(t)
	_, err := commands.Complete(context.Background(), TransitionRequest{
		ReservationID: reservation.ID, ReceiptID: identifierPointer(mustV7(t)),
		CommittedAmount: reservation.Amount + 1,
	})
	if !errors.Is(err, ErrRequest) {
		t.Fatalf("expected invalid amount, got %v", err)
	}
}

package usage

import (
	"context"
	"testing"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func TestCompleteTransitionsAndReplaysExactReceipt(t *testing.T) {
	commands, store, reservation, _ := lifecycleFixture(t)
	receipt := mustV7(t)
	result, err := commands.Complete(context.Background(), TransitionRequest{
		ReservationID: reservation.ID, ReceiptID: identifierPointer(receipt),
	})
	if err != nil || result.Reservation.State != usagedomain.Committed ||
		len(store.transitions) != 1 || store.transitions[0].To != usagedomain.Committed {
		t.Fatalf("result=%+v changes=%+v err=%v", result, store.transitions, err)
	}
	store.locked = result.Reservation
	replay, err := commands.Complete(context.Background(), TransitionRequest{
		ReservationID: reservation.ID, ReceiptID: identifierPointer(receipt),
	})
	if err != nil || !replay.Replay || len(store.transitions) != 1 {
		t.Fatalf("replay=%+v changes=%d err=%v", replay, len(store.transitions), err)
	}
}

func TestCompleteRequiresAuthoritativeReceipt(t *testing.T) {
	commands, _, reservation, _ := lifecycleFixture(t)
	if _, err := commands.Complete(context.Background(), TransitionRequest{
		ReservationID: reservation.ID,
	}); err != ErrRequest {
		t.Fatalf("expected ErrRequest, got %v", err)
	}
}

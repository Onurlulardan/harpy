package usage

import (
	"context"
	"testing"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func TestReleaseBeforeEffectNeedsNoReceipt(t *testing.T) {
	commands, store, reservation, _ := lifecycleFixture(t)
	result, err := commands.Release(context.Background(), TransitionRequest{
		ReservationID: reservation.ID,
	})
	if err != nil || result.Reservation.State != usagedomain.Released ||
		len(store.transitions) != 1 || store.transitions[0].ReceiptID != nil {
		t.Fatalf("result=%+v changes=%+v err=%v", result, store.transitions, err)
	}
}

func TestReleaseAfterEffectRequiresNoEffectReceipt(t *testing.T) {
	commands, store, reservation, now := lifecycleFixture(t)
	effectReceipt := mustV7(t)
	store.locked.EffectStartedAt = instantPointer(now)
	store.locked.EffectReceiptID = identifierPointer(effectReceipt)
	if _, err := commands.Release(context.Background(), TransitionRequest{
		ReservationID: reservation.ID,
	}); err != ErrRequest {
		t.Fatalf("expected ErrRequest, got %v", err)
	}
	proof := mustV7(t)
	result, err := commands.Release(context.Background(), TransitionRequest{
		ReservationID: reservation.ID, ReceiptID: identifierPointer(proof),
	})
	if err != nil || result.Reservation.State != usagedomain.Released {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

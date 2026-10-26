package usage

import (
	"context"
	"testing"
	"time"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func TestExpireUsesEffectEvidenceMatrix(t *testing.T) {
	tests := []struct {
		name     string
		started  bool
		evidence usagedomain.EffectEvidence
		target   usagedomain.ReservationState
		receipt  bool
	}{
		{"not started", false, usagedomain.EffectUnknown, usagedomain.Released, false},
		{"unknown effect", true, usagedomain.EffectUnknown, usagedomain.ReconcileRequired, false},
		{"confirmed effect", true, usagedomain.EffectConfirmed, usagedomain.Committed, true},
		{"confirmed no effect", true, usagedomain.NoEffectConfirmed, usagedomain.Released, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands, store, reservation, now := lifecycleFixture(t)
			store.locked.ExpiresAt = now.Add(-time.Microsecond)
			if test.started {
				effectReceipt := mustV7(t)
				store.locked.EffectStartedAt = instantPointer(now.Add(-time.Minute))
				store.locked.EffectReceiptID = identifierPointer(effectReceipt)
			}
			request := ExpireRequest{ReservationID: reservation.ID, Evidence: test.evidence}
			if test.receipt {
				request.ReceiptID = identifierPointer(mustV7(t))
			}
			result, err := commands.Expire(context.Background(), request)
			if err != nil || result.Reservation.State != test.target || len(store.transitions) != 1 {
				t.Fatalf("result=%+v changes=%+v err=%v", result, store.transitions, err)
			}
		})
	}
}

func TestExpireRejectsBeforeDeadline(t *testing.T) {
	commands, _, reservation, _ := lifecycleFixture(t)
	if _, err := commands.Expire(context.Background(), ExpireRequest{
		ReservationID: reservation.ID, Evidence: usagedomain.EffectUnknown,
	}); err != ErrState {
		t.Fatalf("expected ErrState, got %v", err)
	}
}

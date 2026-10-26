package usage

import (
	"context"
	"time"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func (commands *Commands) resolve(
	ctx context.Context,
	request TransitionRequest,
	target usagedomain.ReservationState,
	requireReceipt bool,
) (ReservationResult, error) {
	now := commands.clock()
	if !validInstant(now) || !validV7(request.ReservationID) ||
		request.ReceiptID != nil && !validV7(*request.ReceiptID) ||
		requireReceipt && request.ReceiptID == nil || request.CommittedAmount < 0 ||
		target != usagedomain.Committed && request.CommittedAmount != 0 {
		return ReservationResult{}, ErrRequest
	}
	execute := commands.store.Complete
	if target == usagedomain.Released {
		execute = commands.store.Release
	}
	result, err := execute(
		ctx, TransitionExecution{Request: request, Now: now},
		func(reservation Reservation, found bool) (ReservationPlan, error) {
			return commands.planResolution(request, now, reservation, found, target)
		},
	)
	if err != nil {
		return ReservationResult{}, usageOperationError("resolve billing usage reservation", err)
	}
	return result, nil
}

func (commands *Commands) planResolution(
	request TransitionRequest,
	now time.Time,
	reservation Reservation,
	found bool,
	target usagedomain.ReservationState,
) (ReservationPlan, error) {
	if !found {
		return ReservationPlan{}, ErrReservationAbsent
	}
	if !validStoredReservation(reservation) || reservation.ID != request.ReservationID {
		return ReservationPlan{}, ErrSnapshot
	}
	if reservation.State == target {
		if !sameReceipt(reservation.ResolutionReceipt, request.ReceiptID) {
			return ReservationPlan{}, ErrReservationReused
		}
		if target == usagedomain.Committed &&
			reservation.CommittedAmount != normalizedCommitAmount(request, reservation) {
			return ReservationPlan{}, ErrReservationReused
		}
		return ReservationPlan{
			Result: ReservationResult{Reservation: reservation, Replay: true},
		}, nil
	}
	if target == usagedomain.Released && reservation.EffectStartedAt != nil &&
		request.ReceiptID == nil {
		return ReservationPlan{}, ErrRequest
	}
	if _, err := usagedomain.Transition(reservation.State, target); err != nil {
		return ReservationPlan{}, ErrState
	}
	committedAmount := int64(0)
	if target == usagedomain.Committed {
		committedAmount = normalizedCommitAmount(request, reservation)
		if committedAmount <= 0 || committedAmount > reservation.Amount {
			return ReservationPlan{}, ErrRequest
		}
	}
	entryID, err := commands.nextID()
	if err != nil || !validV7(entryID) {
		return ReservationPlan{}, ErrConfiguration
	}
	change := ReservationTransition{
		LedgerEntryID: entryID,
		ReservationID: reservation.ID, ExpectedRevision: reservation.Revision,
		From: reservation.State, To: target, ResolvedAt: now,
		ReceiptID: request.ReceiptID, CommittedAmount: committedAmount,
	}
	if target == usagedomain.Committed && committedAmount < reservation.Amount {
		change.RemainderEntryID, err = commands.nextID()
		if err != nil || !validV7(change.RemainderEntryID) ||
			change.RemainderEntryID == change.LedgerEntryID {
			return ReservationPlan{}, ErrConfiguration
		}
	}
	reservation.State = target
	reservation.ResolutionReceipt = request.ReceiptID
	reservation.CommittedAmount = committedAmount
	reservation.Revision++
	return ReservationPlan{
		Result: ReservationResult{Reservation: reservation}, Transition: &change,
	}, nil
}

func normalizedCommitAmount(request TransitionRequest, reservation Reservation) int64 {
	if request.CommittedAmount == 0 {
		return reservation.Amount
	}
	return request.CommittedAmount
}

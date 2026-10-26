package usage

import (
	"context"
	"time"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func (commands *Commands) Expire(
	ctx context.Context,
	request ExpireRequest,
) (ReservationResult, error) {
	now := commands.clock()
	if !validInstant(now) || !validV7(request.ReservationID) || !request.Evidence.Valid() ||
		request.ReceiptID != nil && !validV7(*request.ReceiptID) {
		return ReservationResult{}, ErrRequest
	}
	if (request.Evidence == usagedomain.EffectConfirmed ||
		request.Evidence == usagedomain.NoEffectConfirmed) && request.ReceiptID == nil {
		return ReservationResult{}, ErrRequest
	}
	result, err := commands.store.Expire(
		ctx, ExpireExecution{Request: request, Now: now},
		func(reservation Reservation, found bool) (ReservationPlan, error) {
			return commands.planExpiry(request, now, reservation, found)
		},
	)
	if err != nil {
		return ReservationResult{}, usageOperationError("expire billing usage reservation", err)
	}
	return result, nil
}

func (commands *Commands) planExpiry(
	request ExpireRequest,
	now time.Time,
	reservation Reservation,
	found bool,
) (ReservationPlan, error) {
	if !found {
		return ReservationPlan{}, ErrReservationAbsent
	}
	if !validStoredReservation(reservation) || reservation.ID != request.ReservationID {
		return ReservationPlan{}, ErrSnapshot
	}
	if reservation.State == usagedomain.Committed || reservation.State == usagedomain.Released {
		return ReservationPlan{
			Result: ReservationResult{Reservation: reservation, Replay: true},
		}, nil
	}
	if now.Before(reservation.ExpiresAt) {
		return ReservationPlan{}, ErrState
	}
	target, err := usagedomain.ResolveExpiry(
		reservation.EffectStartedAt != nil, request.Evidence)
	if err != nil {
		return ReservationPlan{}, ErrState
	}
	if reservation.State == target {
		return ReservationPlan{
			Result: ReservationResult{Reservation: reservation, Replay: true},
		}, nil
	}
	entryID, err := commands.nextID()
	if err != nil || !validV7(entryID) {
		return ReservationPlan{}, ErrConfiguration
	}
	change := ReservationTransition{
		LedgerEntryID: entryID,
		ReservationID: reservation.ID, ExpectedRevision: reservation.Revision,
		From: reservation.State, To: target, ResolvedAt: now,
		ReceiptID: request.ReceiptID,
	}
	if target == usagedomain.Committed {
		change.CommittedAmount = reservation.Amount
		reservation.CommittedAmount = reservation.Amount
	}
	reservation.State = target
	reservation.ResolutionReceipt = request.ReceiptID
	reservation.Revision++
	return ReservationPlan{
		Result: ReservationResult{Reservation: reservation}, Transition: &change,
	}, nil
}

package usage

import (
	"context"
	"time"

	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func (commands *Commands) StartEffect(
	ctx context.Context,
	request StartEffectRequest,
) (ReservationResult, error) {
	now := commands.clock()
	if !validInstant(now) || !validV7(request.ReservationID) || !validV7(request.ReceiptID) {
		return ReservationResult{}, ErrRequest
	}
	result, err := commands.store.Start(
		ctx, StartExecution{Request: request, Now: now},
		func(reservation Reservation, found bool) (ReservationPlan, error) {
			return commands.planEffectStart(request, now, reservation, found)
		},
	)
	if err != nil {
		return ReservationResult{}, usageOperationError("start billing usage effect", err)
	}
	return result, nil
}

func (commands *Commands) planEffectStart(
	request StartEffectRequest,
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
	if reservation.EffectStartedAt != nil {
		if *reservation.EffectReceiptID != request.ReceiptID {
			return ReservationPlan{}, ErrReservationReused
		}
		return ReservationPlan{
			Result: ReservationResult{Reservation: reservation, Replay: true},
		}, nil
	}
	if reservation.State != usagedomain.Reserved || !now.Before(reservation.ExpiresAt) {
		return ReservationPlan{}, ErrState
	}
	change := EffectStart{
		ReservationID: reservation.ID, ExpectedRevision: reservation.Revision,
		StartedAt: now, ReceiptID: request.ReceiptID,
	}
	reservation.EffectStartedAt = &change.StartedAt
	reservation.EffectReceiptID = &change.ReceiptID
	reservation.Revision++
	return ReservationPlan{
		Result: ReservationResult{Reservation: reservation}, Effect: &change,
	}, nil
}

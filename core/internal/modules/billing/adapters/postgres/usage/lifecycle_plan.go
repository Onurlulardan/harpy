package usage

import (
	"context"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

func (storage *store) planLifecycle(
	ctx context.Context,
	reservationID uuid.UUID,
	planner usageapp.ReservationPlanner,
) (usageapp.ReservationPlan, error) {
	reservation, found, err := storage.lockReservation(ctx, reservationID)
	if err != nil {
		return usageapp.ReservationPlan{}, err
	}
	plan, err := planner(reservation, found)
	if err != nil {
		return usageapp.ReservationPlan{}, err
	}
	if plan.Effect != nil && plan.Transition != nil {
		return usageapp.ReservationPlan{}, usageapp.ErrSnapshot
	}
	if plan.Effect == nil && plan.Transition == nil {
		if !plan.Result.Replay {
			return usageapp.ReservationPlan{}, usageapp.ErrSnapshot
		}
		return plan, nil
	}
	if plan.Result.Replay || !found ||
		plan.Result.Reservation.ID != reservationID ||
		plan.Result.Reservation.Revision != reservation.Revision+1 {
		return usageapp.ReservationPlan{}, usageapp.ErrSnapshot
	}
	return plan, nil
}

func (storage *store) persistTransition(
	ctx context.Context,
	plan usageapp.ReservationPlan,
) (usageapp.ReservationResult, error) {
	if plan.Transition == nil || plan.Effect != nil ||
		plan.Result.Reservation.ID != plan.Transition.ReservationID {
		return usageapp.ReservationResult{}, usageapp.ErrSnapshot
	}
	if err := storage.writeTransition(ctx, *plan.Transition); err != nil {
		return usageapp.ReservationResult{}, err
	}
	return plan.Result, nil
}

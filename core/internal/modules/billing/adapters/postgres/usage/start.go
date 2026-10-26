package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

func (storage *store) Start(
	ctx context.Context,
	execution usageapp.StartExecution,
	planner usageapp.ReservationPlanner,
) (usageapp.ReservationResult, error) {
	if !validExecution(storage, ctx, execution.Now, planner != nil) {
		return usageapp.ReservationResult{}, usageapp.ErrConfiguration
	}
	plan, err := storage.planLifecycle(ctx, execution.Request.ReservationID, planner)
	if err != nil {
		return usageapp.ReservationResult{}, fmt.Errorf("plan effect start: %w", err)
	}
	if plan.Effect == nil {
		if plan.Transition != nil || !plan.Result.Replay {
			return usageapp.ReservationResult{}, usageapp.ErrSnapshot
		}
		return plan.Result, nil
	}
	if plan.Transition != nil || plan.Result.Reservation.ID != plan.Effect.ReservationID {
		return usageapp.ReservationResult{}, usageapp.ErrSnapshot
	}
	if err := storage.writeEffectStart(ctx, *plan.Effect); err != nil {
		return usageapp.ReservationResult{}, fmt.Errorf("write effect start: %w", err)
	}
	return plan.Result, nil
}

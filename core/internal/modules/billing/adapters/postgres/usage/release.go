package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	usagedomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/usage"
)

func (storage *store) Release(
	ctx context.Context,
	execution usageapp.TransitionExecution,
	planner usageapp.ReservationPlanner,
) (usageapp.ReservationResult, error) {
	if !validExecution(storage, ctx, execution.Now, planner != nil) {
		return usageapp.ReservationResult{}, usageapp.ErrConfiguration
	}
	plan, err := storage.planLifecycle(ctx, execution.Request.ReservationID, planner)
	if err != nil {
		return usageapp.ReservationResult{}, fmt.Errorf("plan release: %w", err)
	}
	if plan.Transition == nil {
		return plan.Result, nil
	}
	if plan.Transition.To != usagedomain.Released {
		return usageapp.ReservationResult{}, usageapp.ErrSnapshot
	}
	result, err := storage.persistTransition(ctx, plan)
	if err != nil {
		return usageapp.ReservationResult{}, fmt.Errorf("write release: %w", err)
	}
	return result, nil
}

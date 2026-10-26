package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

func (storage *store) Expire(
	ctx context.Context,
	execution usageapp.ExpireExecution,
	planner usageapp.ReservationPlanner,
) (usageapp.ReservationResult, error) {
	if !validExecution(storage, ctx, execution.Now, planner != nil) {
		return usageapp.ReservationResult{}, usageapp.ErrConfiguration
	}
	plan, err := storage.planLifecycle(ctx, execution.Request.ReservationID, planner)
	if err != nil {
		return usageapp.ReservationResult{}, fmt.Errorf("plan expiry: %w", err)
	}
	if plan.Transition == nil {
		return plan.Result, nil
	}
	result, err := storage.persistTransition(ctx, plan)
	if err != nil {
		return usageapp.ReservationResult{}, fmt.Errorf("write expiry: %w", err)
	}
	return result, nil
}

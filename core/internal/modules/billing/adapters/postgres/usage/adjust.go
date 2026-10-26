package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

func (storage *store) Adjust(
	ctx context.Context,
	execution usageapp.AdjustExecution,
	planner usageapp.AdjustPlanner,
) (usageapp.AdjustGaugeResult, error) {
	if !validExecution(storage, ctx, execution.Now, planner != nil) {
		return usageapp.AdjustGaugeResult{}, usageapp.ErrConfiguration
	}
	request := execution.Request
	snapshot, err := storage.lockGauge(ctx, request.EntitlementKey, request.PeriodID)
	if err != nil {
		return usageapp.AdjustGaugeResult{}, fmt.Errorf("lock gauge: %w", err)
	}
	replay, found, err := storage.findGaugeEntry(
		ctx, request.EntitlementKey, request.OwnerFactID,
	)
	if err != nil {
		return usageapp.AdjustGaugeResult{}, fmt.Errorf("find gauge entry: %w", err)
	}
	var replayPointer *usageapp.GaugeReplay
	if found {
		replayPointer = &replay
	}
	plan, err := planner(snapshot, replayPointer)
	if err != nil {
		return usageapp.AdjustGaugeResult{}, err
	}
	if plan.Decision == nil {
		if replayPointer == nil || !plan.Result.Replay || plan.Entry != nil ||
			plan.Event != nil {
			return usageapp.AdjustGaugeResult{}, usageapp.ErrSnapshot
		}
		return plan.Result, nil
	}
	if replayPointer != nil || plan.Result.Replay ||
		plan.Result.Allowed != plan.Decision.Allowed ||
		plan.Result.SoftLimitExceeded != plan.Decision.SoftExceeded ||
		!validThresholdEvent(
			plan.Event, plan.Decision, request.PeriodID, snapshot.Revision+1,
			request.OwnerFactID, execution.Now,
		) {
		return usageapp.AdjustGaugeResult{}, usageapp.ErrSnapshot
	}
	if err := storage.writeDecision(ctx, *plan.Decision); err != nil {
		return usageapp.AdjustGaugeResult{}, fmt.Errorf("write decision: %w", err)
	}
	if plan.Entry == nil {
		if plan.Result.Allowed || plan.Result.Entry != (usageapp.GaugeEntry{}) ||
			plan.Event != nil {
			return usageapp.AdjustGaugeResult{}, usageapp.ErrSnapshot
		}
		return plan.Result, nil
	}
	if !plan.Result.Allowed || plan.Result.Entry != *plan.Entry ||
		plan.Entry.DecisionID != plan.Decision.ID {
		return usageapp.AdjustGaugeResult{}, usageapp.ErrSnapshot
	}
	if err := storage.writeGaugeEntry(ctx, *plan.Entry); err != nil {
		return usageapp.AdjustGaugeResult{}, fmt.Errorf("write gauge entry: %w", err)
	}
	if plan.Event != nil {
		if err := storage.effects.PublishThreshold(ctx, *plan.Event); err != nil {
			return usageapp.AdjustGaugeResult{}, fmt.Errorf("publish usage threshold: %w", err)
		}
	}
	return plan.Result, nil
}

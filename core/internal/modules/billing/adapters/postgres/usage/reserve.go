package usage

import (
	"context"
	"fmt"
	"time"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
)

func (storage *store) Reserve(
	ctx context.Context,
	execution usageapp.ReserveExecution,
	planner usageapp.ReservePlanner,
) (usageapp.ReserveResult, error) {
	if !validExecution(storage, ctx, execution.Now, planner != nil) {
		return usageapp.ReserveResult{}, usageapp.ErrConfiguration
	}
	request := execution.Request
	snapshot, err := storage.lockReservationSnapshot(
		ctx, request.EntitlementKey, request.PeriodID,
	)
	if err != nil {
		return usageapp.ReserveResult{}, fmt.Errorf("lock reservation usage: %w", err)
	}
	replay, found, err := storage.findReservation(
		ctx, request.EntitlementKey, request.SourceOperationID,
	)
	if err != nil {
		return usageapp.ReserveResult{}, fmt.Errorf("find reservation: %w", err)
	}
	var replayPointer *usageapp.ReservationReplay
	if found {
		replayPointer = &replay
	}
	plan, err := planner(snapshot, replayPointer)
	if err != nil {
		return usageapp.ReserveResult{}, err
	}
	if plan.Decision == nil {
		if replayPointer == nil || !plan.Result.Replay || plan.Reservation != nil ||
			plan.Event != nil {
			return usageapp.ReserveResult{}, usageapp.ErrSnapshot
		}
		return plan.Result, nil
	}
	if replayPointer != nil || plan.Result.Replay ||
		plan.Result.Allowed != plan.Decision.Allowed ||
		plan.Result.SoftLimitExceeded != plan.Decision.SoftExceeded ||
		!validThresholdEvent(
			plan.Event, plan.Decision, request.PeriodID, snapshot.Revision+1,
			request.SourceOperationID, execution.Now,
		) {
		return usageapp.ReserveResult{}, usageapp.ErrSnapshot
	}
	if err := storage.writeDecision(ctx, *plan.Decision); err != nil {
		return usageapp.ReserveResult{}, fmt.Errorf("write decision: %w", err)
	}
	if plan.Reservation == nil {
		if plan.Result.Allowed || plan.Result.Reservation != (usageapp.Reservation{}) ||
			plan.Event != nil {
			return usageapp.ReserveResult{}, usageapp.ErrSnapshot
		}
		return plan.Result, nil
	}
	if !plan.Result.Allowed || plan.Result.Reservation != *plan.Reservation ||
		plan.Reservation.DecisionID != plan.Decision.ID {
		return usageapp.ReserveResult{}, usageapp.ErrSnapshot
	}
	if err := storage.writeReservation(ctx, *plan.Reservation, execution.Now); err != nil {
		return usageapp.ReserveResult{}, fmt.Errorf("write reservation: %w", err)
	}
	if plan.Event != nil {
		if err := storage.effects.PublishThreshold(ctx, *plan.Event); err != nil {
			return usageapp.ReserveResult{}, fmt.Errorf("publish usage threshold: %w", err)
		}
	}
	return plan.Result, nil
}

func validExecution(storage *store, ctx context.Context, now time.Time, planner bool) bool {
	return storage.usable(ctx) && planner && !now.IsZero() && now.Location() == time.UTC &&
		now.Nanosecond()%1_000 == 0
}

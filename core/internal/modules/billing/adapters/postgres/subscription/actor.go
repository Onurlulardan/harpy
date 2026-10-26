package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

func (storage *store) Actor(
	ctx context.Context,
	execution subscriptionapp.ActorExecution,
	planner subscriptionapp.Planner,
) (subscriptionapp.Result, error) {
	if !storage.usable(ctx, modeActor) || planner == nil ||
		!subscriptionapp.ActorExecutionAccepted(execution) ||
		execution.ActorID != storage.actorID {
		return subscriptionapp.Result{}, subscriptionapp.ErrConfiguration
	}
	authorityID, err := storage.authorityID(ctx, modeActor)
	if err != nil {
		return subscriptionapp.Result{}, err
	}
	locked, receipt, err := storage.loadActor(ctx, authorityID, execution)
	if err != nil {
		return subscriptionapp.Result{}, fmt.Errorf("load actor subscription state: %w", err)
	}
	if receipt != nil && !receipt.matches(execution.Execution, string(modeActor)) {
		return subscriptionapp.Result{}, subscriptionapp.ErrIdempotencyReused
	}
	plan, err := planner(locked)
	if err != nil {
		return subscriptionapp.Result{}, err
	}
	if !validPlan(modeActor, storage.actorID, execution.Execution, locked, plan) ||
		!validActorPlan(execution, plan) {
		return subscriptionapp.Result{}, subscriptionapp.ErrSnapshot
	}
	if !plan.Replay {
		if err := storage.commit(ctx, authorityID, modeActor, plan); err != nil {
			return subscriptionapp.Result{}, fmt.Errorf("commit actor subscription: %w", err)
		}
	}
	return subscriptionapp.Result{Subscription: plan.After, Replay: plan.Replay}, nil
}

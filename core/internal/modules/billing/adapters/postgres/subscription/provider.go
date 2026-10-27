package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

func (storage *store) Provider(
	ctx context.Context,
	execution subscriptionapp.ProviderExecution,
	planner subscriptionapp.Planner,
) (subscriptionapp.Result, error) {
	if !storage.usable(ctx, modeProvider) || planner == nil ||
		!subscriptionapp.ProviderExecutionAccepted(execution) {
		return subscriptionapp.Result{}, subscriptionapp.ErrConfiguration
	}
	authorityID, err := storage.authorityID(ctx, modeProvider)
	if err != nil {
		return subscriptionapp.Result{}, err
	}
	locked, _, err := storage.loadProvider(ctx, authorityID, execution)
	if err != nil {
		return subscriptionapp.Result{}, fmt.Errorf("load provider subscription state: %w", err)
	}
	plan, err := planner(locked)
	if err != nil {
		return subscriptionapp.Result{}, err
	}
	if !validPlan(modeProvider, storage.actorID, execution.Execution, locked, plan) ||
		!validProviderPlan(plan) {
		return subscriptionapp.Result{}, subscriptionapp.ErrSnapshot
	}
	if plan.Replay {
		return subscriptionapp.Result{Subscription: plan.After, Replay: true}, nil
	}
	if err := storage.commit(ctx, authorityID, modeProvider, plan); err != nil {
		return subscriptionapp.Result{}, fmt.Errorf("commit provider subscription: %w", err)
	}
	if err := storage.providerReceiptApplied(ctx, authorityID, execution, plan); err != nil {
		return subscriptionapp.Result{}, err
	}
	return subscriptionapp.Result{Subscription: plan.After}, nil
}

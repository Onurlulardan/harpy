package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

func (storage *store) Job(
	ctx context.Context,
	execution subscriptionapp.JobExecution,
	planner subscriptionapp.Planner,
) (subscriptionapp.Result, error) {
	if !storage.usable(ctx, modeJob) || planner == nil ||
		!subscriptionapp.JobExecutionAccepted(execution) {
		return subscriptionapp.Result{}, subscriptionapp.ErrConfiguration
	}
	authorityID, err := storage.authorityID(ctx, modeJob)
	if err != nil {
		return subscriptionapp.Result{}, err
	}
	locked, _, err := storage.loadJob(ctx, authorityID, execution)
	if err != nil {
		return subscriptionapp.Result{}, fmt.Errorf("load job subscription state: %w", err)
	}
	plan, err := planner(locked)
	if err != nil {
		return subscriptionapp.Result{}, err
	}
	if !validPlan(modeJob, storage.actorID, execution.Execution, locked, plan) ||
		!validJobPlan(execution, plan) || !jobEffectsMatch(execution, plan) {
		return subscriptionapp.Result{}, subscriptionapp.ErrSnapshot
	}
	if !plan.Replay {
		if err := storage.commit(ctx, authorityID, modeJob, plan); err != nil {
			return subscriptionapp.Result{}, fmt.Errorf("commit job subscription: %w", err)
		}
	}
	return subscriptionapp.Result{Subscription: plan.After, Replay: plan.Replay}, nil
}

func jobEffectsMatch(
	execution subscriptionapp.JobExecution,
	plan subscriptionapp.CommitPlan,
) bool {
	if plan.Replay {
		return true
	}
	for _, cancellation := range plan.CancelJobs {
		if cancellation.ID == execution.JobID {
			return false
		}
	}
	if execution.Boundary == subscriptionapp.BoundaryManualChange {
		return plan.ResolveChange != nil
	}
	return plan.ResolveChange == nil
}

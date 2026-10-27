package subscription

import (
	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
)

func validActorPlan(
	execution subscriptionapp.ActorExecution,
	plan subscriptionapp.CommitPlan,
) bool {
	if plan.CompleteJob != nil || plan.ResolveChange != nil {
		return false
	}
	if plan.Replay {
		return plan.Change == nil && plan.TrialClaim == nil
	}
	switch execution.OperationID {
	case subscriptionapp.OperationChangeAtPeriodEnd:
		return plan.Change != nil && plan.TrialClaim == nil
	case subscriptionapp.OperationClaimTrial:
		return plan.Change == nil && plan.TrialClaim != nil
	default:
		return plan.Change == nil && plan.TrialClaim == nil
	}
}

func validProviderPlan(plan subscriptionapp.CommitPlan) bool {
	return plan.CompleteJob == nil && plan.Change == nil &&
		plan.ResolveChange == nil && plan.TrialClaim == nil
}

func validJobPlan(
	execution subscriptionapp.JobExecution,
	plan subscriptionapp.CommitPlan,
) bool {
	if plan.Change != nil || plan.TrialClaim != nil {
		return false
	}
	if plan.Replay {
		return plan.CompleteJob == nil && plan.ResolveChange == nil
	}
	completion := plan.CompleteJob
	return completion != nil && completion.ID == execution.JobID &&
		completion.LeaseGeneration == execution.LeaseGeneration
}

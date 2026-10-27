package subscriptionapp

import (
	"time"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

func (commands *Commands) planPlatformTransition(
	locked LockedState,
	execution ActorExecution,
	request PlatformTransitionRequest,
	reason string,
	now time.Time,
) (CommitPlan, error) {
	if !locked.valid() || locked.Subscription.ID != request.SubscriptionID {
		return CommitPlan{}, ErrSnapshot
	}
	if replay, found, err := receiptReplay(locked.Receipt, execution.Execution); found {
		return replay, err
	}
	if locked.Subscription.Revision != request.ExpectedRevision {
		return CommitPlan{}, ErrStaleRevision
	}
	target := subscriptiondomain.State(request.TargetState)
	if locked.PendingChange != nil || subscriptiondomain.ValidateTransition(
		locked.Subscription.State, target,
	) != nil {
		return CommitPlan{}, ErrStateConflict
	}
	after, err := platformTransitionSnapshot(locked.Subscription, target, now)
	if err != nil {
		return CommitPlan{}, err
	}
	identifiers, err := commands.baseIdentifiers()
	if err != nil {
		return CommitPlan{}, err
	}
	plan := basePlan(planInputs{
		Before: locked.Subscription, After: after, Execution: execution.Execution,
		ReceiptID: identifiers.receipt, AuditID: identifiers.audit,
		EventID: identifiers.subscriptionEvent, EntitlementEventID: identifiers.entitlementEvent,
		ActorID: request.ActorID, Now: now, Kind: "platform_state_transition", Reason: reason,
	})
	plan.CancelJobs = boundaryJobCancellations(locked, now, "platform_state_transition")
	if kind, runAt, found := nextBoundary(after); found {
		if !now.Before(runAt) {
			return CommitPlan{}, ErrStateConflict
		}
		job, jobErr := commands.boundaryJob(after.ID, kind, runAt, after.Revision)
		if jobErr != nil {
			return CommitPlan{}, jobErr
		}
		plan.Jobs = []JobIntent{job}
	}
	if !plan.valid() {
		return CommitPlan{}, ErrConfiguration
	}
	return plan, nil
}

func platformTransitionSnapshot(
	before SubscriptionSnapshot,
	target subscriptiondomain.State,
	now time.Time,
) (SubscriptionSnapshot, error) {
	after := before.revised(now)
	after.State = target
	after.CancelAtPeriodEnd, after.CancelAt = false, nil
	switch target {
	case subscriptiondomain.Trialing:
		if after.Period == nil || after.TrialEndsAt == nil ||
			!after.TrialEndsAt.Equal(after.Period.End()) || !now.Before(*after.TrialEndsAt) {
			return SubscriptionSnapshot{}, ErrStateConflict
		}
		after.GraceEndsAt = nil
	case subscriptiondomain.Active:
		if after.Period == nil || !now.Before(after.Period.End()) {
			return SubscriptionSnapshot{}, ErrStateConflict
		}
		after.TrialEndsAt, after.GraceEndsAt = nil, nil
	case subscriptiondomain.PastDue:
		if after.GraceEndsAt == nil || !now.Before(*after.GraceEndsAt) {
			return SubscriptionSnapshot{}, ErrStateConflict
		}
		after.TrialEndsAt = nil
	case subscriptiondomain.Paused, subscriptiondomain.Cancelled,
		subscriptiondomain.Expired:
		after.TrialEndsAt, after.GraceEndsAt = nil, nil
	default:
		return SubscriptionSnapshot{}, ErrStateConflict
	}
	if !after.valid() {
		return SubscriptionSnapshot{}, ErrStateConflict
	}
	return after, nil
}

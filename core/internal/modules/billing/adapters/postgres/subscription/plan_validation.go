package subscription

import (
	"reflect"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func validPlan(
	mode bindingMode,
	actorID uuid.UUID,
	execution subscriptionapp.Execution,
	locked subscriptionapp.LockedState,
	plan subscriptionapp.CommitPlan,
) bool {
	if !plan.Valid() || plan.Receipt.OperationID != execution.OperationID ||
		plan.Receipt.SourceOperationID != execution.SourceOperationID ||
		plan.Receipt.RequestSHA256 != execution.RequestSHA256 ||
		!reflect.DeepEqual(plan.Receipt.Result, plan.After) {
		return false
	}
	if plan.Replay {
		return locked.Receipt != nil &&
			plan.Receipt.ID == locked.Receipt.ID &&
			reflect.DeepEqual(plan.After, locked.Receipt.Result)
	}
	if locked.Receipt != nil || plan.BeforeRevision != locked.Subscription.Revision ||
		plan.After.ID != locked.Subscription.ID ||
		plan.After.TenantID != locked.Subscription.TenantID ||
		plan.After.ProviderAccountID != locked.Subscription.ProviderAccountID ||
		plan.Audit.SubscriptionID != plan.After.ID ||
		plan.Audit.SourceOperationID != execution.SourceOperationID ||
		plan.Audit.BeforeRevision != plan.BeforeRevision ||
		plan.Audit.AfterRevision != plan.After.Revision {
		return false
	}
	if mode != modeJob && (plan.After.OfferingVersionID != locked.Subscription.OfferingVersionID ||
		plan.After.PriceVersionID != locked.Subscription.PriceVersionID ||
		plan.After.Currency != locked.Subscription.Currency) {
		return false
	}
	expectedActor := uuid.Nil
	expectedEventActor := execution.SourceOperationID
	if mode == modeActor {
		expectedActor = actorID
		expectedEventActor = actorID
	}
	if plan.Audit.ActorID != expectedActor ||
		!validOutbox(plan, expectedEventActor, execution.SourceOperationID) ||
		!validJobs(plan, locked) {
		return false
	}
	return true
}

func validOutbox(
	plan subscriptionapp.CommitPlan,
	actorID, causationID uuid.UUID,
) bool {
	for _, event := range plan.Outbox {
		if event.SubscriptionID != plan.After.ID ||
			event.AggregateRevision != plan.After.Revision ||
			event.ActorID != actorID || event.CausationID != causationID {
			return false
		}
	}
	return true
}

func validJobs(
	plan subscriptionapp.CommitPlan,
	locked subscriptionapp.LockedState,
) bool {
	active := make(map[uuid.UUID]struct{}, len(locked.BoundaryJobs))
	for _, job := range locked.BoundaryJobs {
		active[job.ID] = struct{}{}
	}
	created := map[uuid.UUID]struct{}{}
	for _, job := range plan.Jobs {
		if job.SubscriptionID != plan.After.ID ||
			job.ExpectedRevision != plan.After.Revision {
			return false
		}
		if _, duplicate := created[job.ID]; duplicate {
			return false
		}
		created[job.ID] = struct{}{}
	}
	cancelled := map[uuid.UUID]struct{}{}
	for _, cancellation := range plan.CancelJobs {
		if _, found := active[cancellation.ID]; !found {
			return false
		}
		if _, duplicate := cancelled[cancellation.ID]; duplicate {
			return false
		}
		cancelled[cancellation.ID] = struct{}{}
	}
	return true
}

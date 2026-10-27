package subscriptionapp

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

type ChangeAtPeriodEndRequest struct {
	ActorID                 uuid.UUID
	SubscriptionID          uuid.UUID
	SourceOperationID       uuid.UUID
	TargetOfferingVersionID uuid.UUID
	TargetPriceVersionID    uuid.UUID
	ExpectedRevision        int64
}

func (commands *Commands) ChangeAtPeriodEnd(
	ctx context.Context,
	request ChangeAtPeriodEndRequest,
) (Result, error) {
	now, err := commands.instant()
	if err != nil || ctx == nil || ctx.Err() != nil || !validID(request.ActorID) ||
		!validID(request.SubscriptionID) || !validV7(request.SourceOperationID) ||
		!validID(request.TargetOfferingVersionID) || !validID(request.TargetPriceVersionID) ||
		request.ExpectedRevision < 1 {
		return Result{}, ErrRequest
	}
	digest := changeDigest(request)
	execution := ActorExecution{
		Execution: Execution{
			OperationID: OperationChangeAtPeriodEnd, SubscriptionID: request.SubscriptionID,
			SourceOperationID: request.SourceOperationID,
			ExpectedRevision:  request.ExpectedRevision, RequestSHA256: digest,
		},
		ActorID: request.ActorID, Permission: permissionBillingManage,
		TargetOfferingVersionID: request.TargetOfferingVersionID,
		TargetPriceVersionID:    request.TargetPriceVersionID,
	}
	result, err := commands.store.Actor(ctx, execution, func(locked LockedState) (CommitPlan, error) {
		if !locked.valid() || locked.Subscription.ID != request.SubscriptionID {
			return CommitPlan{}, ErrSnapshot
		}
		if replay, found, replayErr := receiptReplay(locked.Receipt, execution.Execution); found {
			return replay, replayErr
		}
		if locked.TargetPrice == nil || !locked.TargetPrice.valid() {
			return CommitPlan{}, ErrSnapshot
		}
		if locked.Subscription.Revision != request.ExpectedRevision {
			return CommitPlan{}, ErrStaleRevision
		}
		if locked.PendingChange != nil {
			return CommitPlan{}, ErrChangeExists
		}
		target := *locked.TargetPrice
		if target.OfferingVersionID != request.TargetOfferingVersionID ||
			target.PriceVersionID != request.TargetPriceVersionID ||
			(locked.Subscription.OfferingVersionID == target.OfferingVersionID &&
				locked.Subscription.PriceVersionID == target.PriceVersionID) {
			return CommitPlan{}, ErrSnapshot
		}
		grantPeriod, available := locked.Subscription.grantPeriod()
		if !available {
			return CommitPlan{}, ErrStateConflict
		}
		change, changeErr := subscriptiondomain.ScheduleManualChange(
			locked.Subscription.State, grantPeriod, now,
			locked.Subscription.Currency, target.Currency,
			subscriptiondomain.VersionIdentifier(target.OfferingVersionID),
			subscriptiondomain.VersionIdentifier(target.PriceVersionID),
		)
		if changeErr != nil {
			return CommitPlan{}, ErrStateConflict
		}
		after := locked.Subscription.revised(now)
		identifiers, identifierErr := commands.baseIdentifiers()
		if identifierErr != nil {
			return CommitPlan{}, identifierErr
		}
		changeID, identifierErr := commands.identifier()
		if identifierErr != nil {
			return CommitPlan{}, identifierErr
		}
		decisionID, identifierErr := commands.identifier()
		if identifierErr != nil {
			return CommitPlan{}, identifierErr
		}
		job, identifierErr := commands.boundaryJob(
			after.ID, BoundaryManualChange, change.EffectiveAt(), after.Revision,
		)
		if identifierErr != nil {
			return CommitPlan{}, identifierErr
		}
		plan := basePlan(planInputs{
			Before: locked.Subscription, After: after, Execution: execution.Execution,
			ReceiptID: identifiers.receipt, AuditID: identifiers.audit,
			EventID:            identifiers.subscriptionEvent,
			EntitlementEventID: identifiers.entitlementEvent,
			ActorID:            request.ActorID, Now: now,
			Kind: "manual_change_scheduled", Reason: "tenant_manual_change",
		})
		plan.Change = &SubscriptionChange{
			ID: changeID, DecisionID: decisionID, SubscriptionID: after.ID,
			BeforeOfferingVersionID: locked.Subscription.OfferingVersionID,
			BeforePriceVersionID:    locked.Subscription.PriceVersionID,
			AfterOfferingVersionID:  target.OfferingVersionID,
			AfterPriceVersionID:     target.PriceVersionID,
			SourceOperationID:       request.SourceOperationID, State: ChangeScheduled,
			EffectiveAt: change.EffectiveAt(), CreatedAt: now, Revision: 1,
		}
		plan.Jobs = []JobIntent{job}
		plan.CancelJobs = boundaryJobCancellations(
			locked, now, "subscription_manual_change_scheduled",
		)
		if !plan.valid() {
			return CommitPlan{}, ErrConfiguration
		}
		return plan, nil
	})
	if err != nil {
		return Result{}, fmt.Errorf("change billing subscription at period end: %w", err)
	}
	return checkedResult(result, request.SubscriptionID)
}

func changeDigest(request ChangeAtPeriodEndRequest) [32]byte {
	writer := newDigestWriter()
	writer.text(OperationChangeAtPeriodEnd)
	writer.identifier(request.ActorID)
	writer.identifier(request.SubscriptionID)
	writer.identifier(request.SourceOperationID)
	writer.identifier(request.TargetOfferingVersionID)
	writer.identifier(request.TargetPriceVersionID)
	writer.integer(request.ExpectedRevision)
	return writer.sum()
}

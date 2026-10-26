package subscriptionapp

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

type CancelAtPeriodEndRequest struct {
	ActorID           uuid.UUID
	SubscriptionID    uuid.UUID
	SourceOperationID uuid.UUID
	ExpectedRevision  int64
	Reason            string
}

func (commands *Commands) CancelAtPeriodEnd(
	ctx context.Context,
	request CancelAtPeriodEndRequest,
) (Result, error) {
	now, err := commands.instant()
	reason, reasonValid := canonicalOptionalReason(request.Reason)
	if err != nil || ctx == nil || ctx.Err() != nil || !validID(request.ActorID) ||
		!validID(request.SubscriptionID) || !validV7(request.SourceOperationID) ||
		request.ExpectedRevision < 1 || !reasonValid {
		return Result{}, ErrRequest
	}
	digest := cancellationDigest(request, reason)
	execution := ActorExecution{
		Execution: Execution{
			OperationID: OperationCancelAtPeriodEnd, SubscriptionID: request.SubscriptionID,
			SourceOperationID: request.SourceOperationID,
			ExpectedRevision:  request.ExpectedRevision, RequestSHA256: digest,
		},
		ActorID: request.ActorID, Permission: permissionBillingManage,
	}
	result, err := commands.store.Actor(ctx, execution, func(locked LockedState) (CommitPlan, error) {
		if !locked.valid() || locked.Subscription.ID != request.SubscriptionID {
			return CommitPlan{}, ErrSnapshot
		}
		if replay, found, replayErr := receiptReplay(locked.Receipt, execution.Execution); found {
			return replay, replayErr
		}
		if locked.Subscription.Revision != request.ExpectedRevision {
			return CommitPlan{}, ErrStaleRevision
		}
		if locked.Subscription.CancelAtPeriodEnd {
			return CommitPlan{}, ErrCancellationExists
		}
		grantPeriod, available := locked.Subscription.grantPeriod()
		if !available {
			return CommitPlan{}, ErrStateConflict
		}
		schedule, scheduleErr := subscriptiondomain.ScheduleCancellation(
			locked.Subscription.State, grantPeriod, now,
		)
		if scheduleErr != nil {
			return CommitPlan{}, ErrStateConflict
		}
		after := locked.Subscription.revised(now)
		effectiveAt := schedule.EffectiveAt()
		after.CancelAtPeriodEnd, after.CancelAt = true, &effectiveAt
		identifiers, identifierErr := commands.baseIdentifiers()
		if identifierErr != nil {
			return CommitPlan{}, identifierErr
		}
		job, jobErr := commands.boundaryJob(
			after.ID, BoundaryCancellation, effectiveAt, after.Revision,
		)
		if jobErr != nil {
			return CommitPlan{}, jobErr
		}
		plan := basePlan(planInputs{
			Before: locked.Subscription, After: after, Execution: execution.Execution,
			ReceiptID: identifiers.receipt, AuditID: identifiers.audit,
			EventID:            identifiers.subscriptionEvent,
			EntitlementEventID: identifiers.entitlementEvent,
			ActorID:            request.ActorID, Now: now,
			Kind: "cancellation_scheduled", Reason: reason,
		})
		plan.Jobs = []JobIntent{job}
		plan.CancelJobs = boundaryJobCancellations(
			locked, now, "subscription_cancellation_scheduled",
		)
		if !plan.valid() {
			return CommitPlan{}, ErrConfiguration
		}
		return plan, nil
	})
	if err != nil {
		return Result{}, fmt.Errorf("cancel billing subscription at period end: %w", err)
	}
	return checkedResult(result, request.SubscriptionID)
}

func cancellationDigest(request CancelAtPeriodEndRequest, reason string) [32]byte {
	writer := newDigestWriter()
	writer.text(OperationCancelAtPeriodEnd)
	writer.identifier(request.ActorID)
	writer.identifier(request.SubscriptionID)
	writer.identifier(request.SourceOperationID)
	writer.integer(request.ExpectedRevision)
	writer.text(reason)
	return writer.sum()
}

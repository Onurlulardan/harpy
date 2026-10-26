package subscriptionapp

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

type ResumeCancellationRequest struct {
	ActorID           uuid.UUID
	SubscriptionID    uuid.UUID
	SourceOperationID uuid.UUID
	ExpectedRevision  int64
}

func (commands *Commands) ResumeCancellation(
	ctx context.Context,
	request ResumeCancellationRequest,
) (Result, error) {
	now, err := commands.instant()
	if err != nil || ctx == nil || ctx.Err() != nil || !validID(request.ActorID) ||
		!validID(request.SubscriptionID) || !validV7(request.SourceOperationID) ||
		request.ExpectedRevision < 1 {
		return Result{}, ErrRequest
	}
	digest := resumeDigest(request)
	execution := ActorExecution{
		Execution: Execution{
			OperationID: OperationResumeCancellation, SubscriptionID: request.SubscriptionID,
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
		cancellationJob, jobFound := locked.boundaryJob(BoundaryCancellation)
		if !locked.Subscription.CancelAtPeriodEnd || locked.Subscription.CancelAt == nil ||
			!jobFound {
			return CommitPlan{}, ErrCancellationMissing
		}
		if locked.Subscription.State != subscriptiondomain.Active &&
			locked.Subscription.State != subscriptiondomain.Trialing ||
			!now.Before(*locked.Subscription.CancelAt) ||
			!cancellationJob.RunAt.Equal(*locked.Subscription.CancelAt) {
			return CommitPlan{}, ErrStateConflict
		}
		after := locked.Subscription.revised(now)
		after.CancelAtPeriodEnd, after.CancelAt = false, nil
		identifiers, identifierErr := commands.baseIdentifiers()
		if identifierErr != nil {
			return CommitPlan{}, identifierErr
		}
		plan := basePlan(planInputs{
			Before: locked.Subscription, After: after, Execution: execution.Execution,
			ReceiptID: identifiers.receipt, AuditID: identifiers.audit,
			EventID:            identifiers.subscriptionEvent,
			EntitlementEventID: identifiers.entitlementEvent,
			ActorID:            request.ActorID, Now: now,
			Kind: "cancellation_resumed", Reason: "tenant_resume",
		})
		job, identifierErr := commands.boundaryJob(
			after.ID, BoundaryPeriodEnd, locked.Subscription.Period.End(), after.Revision,
		)
		if identifierErr != nil {
			return CommitPlan{}, identifierErr
		}
		plan.Jobs = []JobIntent{job}
		plan.CancelJobs = []JobCancellation{{
			ID: cancellationJob.ID, CancelledAt: now,
			Reason: "subscription_cancellation_resumed",
		}}
		if !plan.valid() {
			return CommitPlan{}, ErrConfiguration
		}
		return plan, nil
	})
	if err != nil {
		return Result{}, fmt.Errorf("resume billing subscription cancellation: %w", err)
	}
	return checkedResult(result, request.SubscriptionID)
}

func resumeDigest(request ResumeCancellationRequest) [32]byte {
	writer := newDigestWriter()
	writer.text(OperationResumeCancellation)
	writer.identifier(request.ActorID)
	writer.identifier(request.SubscriptionID)
	writer.identifier(request.SourceOperationID)
	writer.integer(request.ExpectedRevision)
	return writer.sum()
}

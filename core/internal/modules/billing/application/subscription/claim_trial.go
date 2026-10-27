package subscriptionapp

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

type ClaimTrialRequest struct {
	ActorID           uuid.UUID
	SubscriptionID    uuid.UUID
	OfferingVersionID uuid.UUID
	SourceOperationID uuid.UUID
	ExpectedRevision  int64
}

func (commands *Commands) ClaimTrial(
	ctx context.Context,
	request ClaimTrialRequest,
) (Result, error) {
	now, err := commands.instant()
	if err != nil || ctx == nil || ctx.Err() != nil || !validID(request.ActorID) ||
		!validID(request.SubscriptionID) || !validID(request.OfferingVersionID) ||
		!validV7(request.SourceOperationID) || request.ExpectedRevision < 1 {
		return Result{}, ErrRequest
	}
	digest := claimTrialDigest(request)
	execution := ActorExecution{
		Execution: Execution{
			OperationID: OperationClaimTrial, SubscriptionID: request.SubscriptionID,
			SourceOperationID: request.SourceOperationID,
			ExpectedRevision:  request.ExpectedRevision, RequestSHA256: digest,
		},
		ActorID: request.ActorID, Permission: permissionBillingManage,
		TargetOfferingVersionID: request.OfferingVersionID,
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
		if locked.TrialClaim != nil {
			return CommitPlan{}, ErrTrialClaimed
		}
		policy := *locked.TargetPrice
		if locked.Subscription.State != subscriptiondomain.Pending || policy.TrialDays < 1 ||
			policy.OfferingVersionID != request.OfferingVersionID ||
			policy.OfferingVersionID != locked.Subscription.OfferingVersionID ||
			policy.PriceVersionID != locked.Subscription.PriceVersionID ||
			policy.Currency != locked.Subscription.Currency ||
			subscriptiondomain.ValidateTransition(
				locked.Subscription.State, subscriptiondomain.Trialing) != nil {
			return CommitPlan{}, ErrStateConflict
		}
		trialEnd := now.AddDate(0, 0, policy.TrialDays)
		trialPeriod, periodErr := billingperiod.New(now, trialEnd)
		if periodErr != nil || !validInstant(trialEnd) {
			return CommitPlan{}, ErrConfiguration
		}
		after := locked.Subscription.revised(now)
		after.State, after.Period = subscriptiondomain.Trialing, &trialPeriod
		after.TrialEndsAt, after.GraceEndsAt = &trialEnd, nil
		after.CancelAtPeriodEnd, after.CancelAt = false, nil
		identifiers, identifierErr := commands.baseIdentifiers()
		if identifierErr != nil {
			return CommitPlan{}, identifierErr
		}
		claimID, identifierErr := commands.identifier()
		if identifierErr != nil {
			return CommitPlan{}, identifierErr
		}
		job, identifierErr := commands.boundaryJob(
			after.ID, BoundaryTrialEnd, trialEnd, after.Revision,
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
			Kind: "trial_claimed", Reason: "tenant_trial_claim",
		})
		plan.TrialClaim = &TrialClaim{
			ID: claimID, SubscriptionID: after.ID,
			OfferingVersionID: request.OfferingVersionID,
			SourceOperationID: request.SourceOperationID, ClaimedAt: now,
		}
		plan.Jobs = []JobIntent{job}
		if !plan.valid() {
			return CommitPlan{}, ErrConfiguration
		}
		return plan, nil
	})
	if err != nil {
		return Result{}, fmt.Errorf("claim billing subscription trial: %w", err)
	}
	return checkedResult(result, request.SubscriptionID)
}

func claimTrialDigest(request ClaimTrialRequest) [32]byte {
	writer := newDigestWriter()
	writer.text(OperationClaimTrial)
	writer.identifier(request.ActorID)
	writer.identifier(request.SubscriptionID)
	writer.identifier(request.OfferingVersionID)
	writer.identifier(request.SourceOperationID)
	writer.integer(request.ExpectedRevision)
	return writer.sum()
}

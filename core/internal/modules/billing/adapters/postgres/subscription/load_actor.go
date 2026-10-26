package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) loadActor(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.ActorExecution,
) (subscriptionapp.LockedState, *receiptRecord, error) {
	locked, receipt, err := storage.loadBase(ctx, authorityID, execution.Execution)
	if err != nil || receipt != nil {
		return locked, receipt, err
	}
	jobs, err := storage.listBoundaryJobs(
		ctx, authorityID, execution.SubscriptionID,
	)
	if err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	locked.BoundaryJobs = jobs
	pending, err := storage.findPendingChange(ctx, authorityID, execution.SubscriptionID)
	if err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	locked.PendingChange = pending
	if err := storage.loadActorTarget(ctx, authorityID, &locked, execution); err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	return locked, nil, nil
}

func (storage *store) loadActorTarget(
	ctx context.Context,
	authorityID uuid.UUID,
	locked *subscriptionapp.LockedState,
	execution subscriptionapp.ActorExecution,
) error {
	if execution.TargetOfferingVersionID == uuid.Nil {
		return nil
	}
	priceID := execution.TargetPriceVersionID
	if priceID == uuid.Nil {
		priceID = locked.Subscription.PriceVersionID
	}
	target, err := storage.findTargetPrice(
		ctx, execution.TargetOfferingVersionID, priceID,
	)
	if err != nil {
		return err
	}
	if target != nil {
		locked.TargetPrice = &target.price
	}
	if execution.OperationID != subscriptionapp.OperationClaimTrial {
		return nil
	}
	claim, err := storage.findTrialClaim(
		ctx, authorityID, execution.TargetOfferingVersionID,
	)
	if err != nil {
		return fmt.Errorf("load actor trial claim: %w", err)
	}
	locked.TrialClaim = claim
	return nil
}

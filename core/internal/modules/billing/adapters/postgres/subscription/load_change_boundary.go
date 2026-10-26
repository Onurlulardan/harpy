package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	billingperiod "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/period"
	subscriptiondomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
	"github.com/google/uuid"
)

func (storage *store) loadChangeBoundary(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.JobExecution,
	job assertedJob,
	state subscriptiondomain.State,
	boundary *subscriptionapp.BoundarySnapshot,
) (*subscriptionapp.BoundarySnapshot, *subscriptionapp.SubscriptionChange, error) {
	change, err := storage.findPendingChange(
		ctx, authorityID, execution.SubscriptionID,
	)
	if err != nil || change == nil {
		if err == nil {
			err = subscriptionapp.ErrBoundaryMismatch
		}
		return nil, nil, err
	}
	target, err := storage.findPriceVersion(
		ctx, change.AfterOfferingVersionID, change.AfterPriceVersionID,
	)
	if err != nil || target == nil {
		if err == nil {
			err = subscriptionapp.ErrBoundaryMismatch
		}
		return nil, nil, err
	}
	boundary.Change, boundary.TargetPrice = change, &target.price
	if state == subscriptiondomain.Active {
		anchor, anchorErr := billingperiod.AnchorFrom(job.runAt)
		if anchorErr != nil {
			return nil, nil, fmt.Errorf("build changed billing anchor: %w", anchorErr)
		}
		next, nextErr := billingperiod.BuildManual(job.runAt, anchor, target.cadence)
		if nextErr != nil {
			return nil, nil, fmt.Errorf("build changed billing period: %w", nextErr)
		}
		boundary.NextPeriod = &next
		return boundary, change, nil
	}
	evidence, err := storage.findBoundaryEvidence(ctx, authorityID, execution, job)
	if err != nil {
		return nil, nil, err
	}
	if err := evidence.apply(boundary); err != nil {
		return nil, nil, err
	}
	return boundary, change, nil
}

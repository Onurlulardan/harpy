package subscription

import (
	"context"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) loadProvider(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.ProviderExecution,
) (subscriptionapp.LockedState, *receiptRecord, error) {
	locked, receipt, err := storage.loadBase(ctx, authorityID, execution.Execution)
	if err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	if receipt != nil && !receipt.matches(execution.Execution, string(modeProvider)) {
		return subscriptionapp.LockedState{}, nil, subscriptionapp.ErrProviderFactReused
	}
	if err := storage.assertProviderFact(
		ctx, authorityID, execution, receipt != nil,
	); err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	if receipt != nil {
		return locked, receipt, nil
	}
	jobs, err := storage.listBoundaryJobs(
		ctx, authorityID, execution.SubscriptionID,
	)
	if err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	locked.BoundaryJobs = jobs
	return locked, nil, nil
}

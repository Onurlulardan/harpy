package subscription

import (
	"context"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) loadBase(
	ctx context.Context,
	authorityID uuid.UUID,
	execution subscriptionapp.Execution,
) (subscriptionapp.LockedState, *receiptRecord, error) {
	snapshot, err := storage.lockSubscription(ctx, authorityID, execution.SubscriptionID)
	if err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	if snapshot.ID != execution.SubscriptionID || snapshot.TenantID != storage.tenantID {
		return subscriptionapp.LockedState{}, nil, subscriptionapp.ErrSnapshot
	}
	receipt, err := storage.findReceipt(ctx, authorityID, execution)
	if err != nil {
		return subscriptionapp.LockedState{}, nil, err
	}
	locked := subscriptionapp.LockedState{Subscription: snapshot}
	if receipt != nil {
		locked.Receipt = &receipt.receipt
	}
	return locked, receipt, nil
}

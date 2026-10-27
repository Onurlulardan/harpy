package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *queryStore) GetCurrent(
	ctx context.Context,
) (subscriptionapp.CurrentSubscription, error) {
	if !storage.queryUsable(ctx) {
		return subscriptionapp.CurrentSubscription{}, subscriptionapp.ErrQueryConfiguration
	}
	authorityID, err := storage.snapshotAuthority(ctx)
	if err != nil {
		return subscriptionapp.CurrentSubscription{}, err
	}
	record, err := scanCurrent(storage.queryRow(
		ctx, getCurrentSQL, authorityID, storage.tenantID,
	))
	if err != nil {
		if storage.notFound(err) {
			return subscriptionapp.CurrentSubscription{}, subscriptionapp.ErrCurrentSubscriptionNotFound
		}
		return subscriptionapp.CurrentSubscription{}, fmt.Errorf(
			"%w: read current subscription", subscriptionapp.ErrSubscriptionQueryStorage,
		)
	}
	return record.value()
}

func (storage *queryStore) snapshotAuthority(
	ctx context.Context,
) (uuid.UUID, error) {
	var authorityID uuid.UUID
	err := storage.queryRow(
		ctx, resolveSnapshotAuthoritySQL, storage.tenantID,
	).Scan(&authorityID)
	if err != nil {
		if storage.notFound(err) {
			return uuid.Nil, subscriptionapp.ErrSubscriptionQueryDependency
		}
		return uuid.Nil, fmt.Errorf(
			"%w: resolve current authority", subscriptionapp.ErrSubscriptionQueryStorage,
		)
	}
	if authorityID == uuid.Nil {
		return uuid.Nil, subscriptionapp.ErrSubscriptionQueryDependency
	}
	return authorityID, nil
}

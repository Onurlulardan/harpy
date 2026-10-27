package usage

import (
	"context"
	"fmt"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

type currentContext struct {
	authorityID                                   uuid.UUID
	baseOfferingVersionID                         uuid.UUID
	subscriptionID, subscriptionOfferingVersionID *uuid.UUID
}

func (storage *queryStore) currentContext(
	ctx context.Context,
) (currentContext, error) {
	authorityID, err := storage.snapshotAuthority(ctx)
	if err != nil {
		return currentContext{}, err
	}
	var baseCount, subscriptionCount int64
	var base, subscription, subscriptionOffering dbText
	err = storage.queryRow(
		ctx, resolveCurrentContextSQL, authorityID, storage.tenantID,
	).Scan(&baseCount, &base, &subscriptionCount, &subscription, &subscriptionOffering)
	if err != nil {
		return currentContext{}, fmt.Errorf("%w: resolve usage context", usageapp.ErrQueryStorage)
	}
	if baseCount == 0 {
		return currentContext{}, usageapp.ErrQueryDependency
	}
	if baseCount != 1 || !base.Valid || subscriptionCount < 0 || subscriptionCount > 1 {
		return currentContext{}, usageapp.ErrQueryInvariant
	}
	baseID, err := uuid.Parse(base.Text)
	if err != nil || baseID == uuid.Nil {
		return currentContext{}, usageapp.ErrQueryInvariant
	}
	result := currentContext{authorityID: authorityID, baseOfferingVersionID: baseID}
	if subscriptionCount == 0 {
		if subscription.Valid || subscriptionOffering.Valid {
			return currentContext{}, usageapp.ErrQueryInvariant
		}
		return result, nil
	}
	if !subscription.Valid || !subscriptionOffering.Valid {
		return currentContext{}, usageapp.ErrQueryInvariant
	}
	subscriptionID, firstErr := uuid.Parse(subscription.Text)
	offeringID, secondErr := uuid.Parse(subscriptionOffering.Text)
	if firstErr != nil || secondErr != nil || subscriptionID == uuid.Nil || offeringID == uuid.Nil {
		return currentContext{}, usageapp.ErrQueryInvariant
	}
	result.subscriptionID = &subscriptionID
	result.subscriptionOfferingVersionID = &offeringID
	return result, nil
}

func (storage *queryStore) snapshotAuthority(ctx context.Context) (uuid.UUID, error) {
	var authorityID uuid.UUID
	err := storage.queryRow(ctx, resolveSnapshotAuthoritySQL, storage.tenantID).Scan(&authorityID)
	if err != nil {
		if storage.notFound(err) {
			return uuid.Nil, usageapp.ErrQueryDependency
		}
		return uuid.Nil, fmt.Errorf("%w: resolve usage authority", usageapp.ErrQueryStorage)
	}
	if authorityID == uuid.Nil {
		return uuid.Nil, usageapp.ErrQueryDependency
	}
	return authorityID, nil
}

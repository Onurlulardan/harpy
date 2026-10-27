package entitlement

import (
	"context"
	"fmt"

	domain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/subscription"
)

func (queries *Queries) Resolve(ctx context.Context) (Resolution, error) {
	now := queries.clock()
	if !validInstant(now) {
		return Resolution{}, ErrConfiguration
	}
	snapshot, err := queries.store.LoadEffectiveSnapshot(ctx, now)
	if err != nil {
		return Resolution{}, fmt.Errorf("load billing entitlements: %w", err)
	}
	if !snapshot.valid() {
		return Resolution{}, ErrSnapshot
	}
	var purchasedValues map[domain.Key]domain.Value
	var overrides map[domain.Key]domain.Value
	result := Resolution{
		baseAssignmentID:      snapshot.BaseAssignmentID,
		baseOfferingVersionID: snapshot.BaseOfferingVersionID,
	}
	if snapshot.Purchased != nil {
		active, err := subscription.PurchasedGrantActive(
			snapshot.Purchased.State, snapshot.Purchased.GrantWindow, now)
		if err != nil {
			return Resolution{}, ErrSnapshot
		}
		if active {
			purchasedValues = snapshot.Purchased.Values
			overrides = snapshot.Purchased.Overrides
			result.subscriptionID = copyIdentifier(&snapshot.Purchased.SubscriptionID)
			result.offeringVersionID = copyIdentifier(&snapshot.Purchased.OfferingVersionID)
		}
	}
	values, err := domain.Resolve(
		snapshot.Definitions, snapshot.Base, purchasedValues, overrides)
	if err != nil {
		return Resolution{}, ErrSnapshot
	}
	result.values = values
	return result, nil
}

package offering

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
	"github.com/google/uuid"
)

func (storage *store) addChildren(
	ctx context.Context,
	authorityID uuid.UUID,
	versionID uuid.UUID,
	priceIDs []uuid.UUID,
	version offeringdomain.Version,
	occurredAt time.Time,
) error {
	prices := version.Prices()
	if len(prices) != len(priceIDs) {
		return fmt.Errorf("billing offering price identifiers do not match prices")
	}
	for index, price := range prices {
		amount := price.Amount()
		if err := storage.execute(ctx, addPriceSQL, authorityID, versionID,
			priceIDs[index], string(price.Cadence()), amount.MinorUnits(),
			amount.Currency(), amount.CurrencyRegistryVersion(), occurredAt); err != nil {
			return fmt.Errorf("add billing offering price: %w", err)
		}
	}
	for _, grant := range version.Grants() {
		booleanValue, integerValue := entitlementArguments(grant.Value())
		if err := storage.execute(ctx, addGrantSQL, authorityID, versionID,
			string(grant.Definition().Key()), string(grant.Value().Kind()),
			booleanValue, integerValue, grant.Value().Unlimited(), occurredAt); err != nil {
			return fmt.Errorf("add billing offering grant: %w", err)
		}
	}
	return nil
}

func entitlementArguments(value entitlement.Value) (any, any) {
	if booleanValue, present := value.Boolean(); present {
		return booleanValue, nil
	}
	if integerValue, present := value.Integer(); present {
		return nil, integerValue
	}
	return nil, nil
}

package discount

import (
	"context"
	"fmt"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	"github.com/google/uuid"
)

func (storage *store) findRedemption(
	ctx context.Context,
	sourceOperationID uuid.UUID,
) (discountapp.Redemption, bool, error) {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return discountapp.Redemption{}, false, err
	}
	result := discountapp.Redemption{}
	var policySHA256 []byte
	err = storage.queryRow(ctx, findRedemptionSQL, authorityID, storage.tenantID,
		sourceOperationID,
	).Scan(&result.ID, &result.DiscountVersionID, &result.SubscriptionID,
		&result.OfferingVersionID, &result.PriceVersionID, &result.SourceOperationID,
		&policySHA256, &result.RedeemedAt)
	if storage.notFound(err) {
		return discountapp.Redemption{}, false, nil
	}
	if err != nil {
		return discountapp.Redemption{}, false,
			fmt.Errorf("find billing discount redemption: %w", err)
	}
	if len(policySHA256) != len(result.PolicySHA256) {
		return discountapp.Redemption{}, false, discountapp.ErrSnapshot
	}
	copy(result.PolicySHA256[:], policySHA256)
	return result, true, nil
}

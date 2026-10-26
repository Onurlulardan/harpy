package discount

import (
	"context"
	"fmt"

	discountapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/discount"
	"github.com/google/uuid"
)

func (storage *store) createRedemption(
	ctx context.Context,
	redemption discountapp.Redemption,
) error {
	authorityID, err := storage.authorityID(ctx)
	if err != nil {
		return err
	}
	var insertedID uuid.UUID
	err = storage.queryRow(ctx, createRedemptionSQL, authorityID, storage.tenantID,
		redemption.ID, redemption.SubscriptionID, redemption.DiscountVersionID,
		redemption.OfferingVersionID, redemption.PriceVersionID,
		redemption.SourceOperationID, redemption.PolicySHA256[:], redemption.RedeemedAt,
	).Scan(&insertedID)
	if err != nil {
		return fmt.Errorf("create billing discount redemption: %w", err)
	}
	if insertedID != redemption.ID {
		return discountapp.ErrSnapshot
	}
	return nil
}

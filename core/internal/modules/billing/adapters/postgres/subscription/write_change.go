package subscription

import (
	"context"
	"fmt"

	subscriptionapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/subscription"
	"github.com/google/uuid"
)

func (storage *store) resolveChange(
	ctx context.Context,
	authorityID, subscriptionID uuid.UUID,
	resolution *subscriptionapp.ChangeResolution,
) error {
	if resolution == nil {
		return nil
	}
	err := storage.expectID(
		ctx, resolveChangeSQL, resolution.ID,
		authorityID, storage.tenantID, resolution.ID, subscriptionID,
		resolution.ExpectedRevision, resolution.ResolvedAt,
	)
	if err != nil {
		return fmt.Errorf("resolve billing subscription change: %w", err)
	}
	return nil
}

func (storage *store) writeChange(
	ctx context.Context,
	authorityID uuid.UUID,
	change *subscriptionapp.SubscriptionChange,
) error {
	if change == nil {
		return nil
	}
	err := storage.expectID(
		ctx, insertChangeSQL, change.ID,
		authorityID, storage.tenantID, change.ID, change.SubscriptionID,
		change.BeforeOfferingVersionID, change.BeforePriceVersionID,
		change.AfterOfferingVersionID, change.AfterPriceVersionID,
		change.EffectiveAt, change.SourceOperationID, change.DecisionID,
		change.CreatedAt, change.Revision,
	)
	if err != nil {
		return fmt.Errorf("insert billing subscription change: %w", err)
	}
	return nil
}
